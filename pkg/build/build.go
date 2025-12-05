// MIT License
//
// Copyright (c) 2018 buildtool
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package build

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	dockerbuild "github.com/docker/docker/api/types/build"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stringid"
	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/filesync"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/opencontainers/go-digest"
	"github.com/tonistiigi/fsutil"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"

	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/ci"
	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/docker"
)

type Args struct {
	args.Globals
	Dockerfile string   `name:"file" short:"f" help:"name of the Dockerfile to use, or '-' to read from stdin" default:"Dockerfile"`
	BuildArgs  []string `name:"build-arg" type:"list" help:"additional docker build-args to use, see https://docs.docker.com/engine/reference/commandline/build/ for more information."`
	NoLogin    bool     `help:"disable login to docker registry" default:"false" `
	NoPull     bool     `help:"disable pulling latest from docker registry" default:"false"`
	Platform   string   `help:"specify target platform(s) to build (e.g. 'linux/amd64' or 'linux/amd64,linux/arm64' for multi-platform)" default:""`
}

// BuildkitClient defines the interface for buildkit operations.
// This interface allows for mocking in tests.
type BuildkitClient interface {
	Solve(ctx context.Context, def *llb.Definition, opt client.SolveOpt, statusChan chan *client.SolveStatus) (*client.SolveResponse, error)
	ListWorkers(ctx context.Context, opts ...client.ListWorkersOption) ([]*client.WorkerInfo, error)
	Close() error
}

// BuildkitClientFactory creates BuildkitClient instances.
// This is used to inject mock clients for testing.
type BuildkitClientFactory func(ctx context.Context, address string, opts ...client.ClientOpt) (BuildkitClient, error)

// defaultBuildkitClientFactory creates real buildkit clients.
var defaultBuildkitClientFactory BuildkitClientFactory = func(ctx context.Context, address string, opts ...client.ClientOpt) (BuildkitClient, error) {
	return client.New(ctx, address, opts...)
}

func (a Args) isDockerfileFromStdin() bool {
	return a.Dockerfile == "-"
}

func (a Args) dockerfileName() string {
	if a.isDockerfileFromStdin() {
		return ""
	}
	return a.Dockerfile
}

func (a Args) isMultiPlatform() bool {
	return a.Platform != "" && strings.Contains(a.Platform, ",")
}

func (a Args) platformCount() int {
	if a.Platform == "" {
		return 0
	}
	return len(strings.Split(a.Platform, ","))
}

// extractHost extracts the host portion from a registry URL.
// For example, "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache-repo" returns "123456789012.dkr.ecr.us-east-1.amazonaws.com".
func extractHost(url string) string {
	// Remove any protocol prefix
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Split by "/" and return the first part (the host)
	parts := strings.SplitN(url, "/", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return url
}

func DoBuild(dir string, buildArgs Args) error {
	dkrClient, err := dockerClient()
	if err != nil {
		return err
	}
	return build(dkrClient, dir, buildArgs)
}

var dockerClient = docker.DefaultClient

var setupSession = provideSession

func provideSession(dir string) Session {
	s, err := session.NewSession(context.Background(), getBuildSharedKey(dir))
	if err != nil {
		panic("session.NewSession changed behaviour and returned an error. Create an issue at https://github.com/buildtool/build-tools/issues/new")
	}
	if s == nil {
		panic("session.NewSession changed behaviour and did not return a session. Create an issue at https://github.com/buildtool/build-tools/issues/new")
	}
	return s
}

func build(client docker.Client, dir string, buildVars Args) error {
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	currentCI := cfg.CurrentCI()
	if buildVars.Platform != "" {
		platforms := strings.Split(buildVars.Platform, ",")
		if len(platforms) > 1 {
			log.Infof("building for <cyan>%d</cyan> platforms: <green>%s</green>\n", len(platforms), buildVars.Platform)
		} else {
			log.Infof("building for platform <green>%s</green>\n", buildVars.Platform)
		}
	}

	log.Debugf("Using CI <green>%s</green>\n", currentCI.Name())

	currentRegistry := cfg.CurrentRegistry()
	log.Debugf("Using registry <green>%s</green>\n", currentRegistry.Name())
	var authenticator docker.Authenticator
	if buildVars.NoLogin {
		log.Debugf("Login <yellow>disabled</yellow>\n")
	} else {
		log.Debugf("Authenticating against registry <green>%s</green>\n", currentRegistry.Name())
		if err := currentRegistry.Login(client); err != nil {
			return err
		}
		authenticator = docker.NewAuthenticator(currentRegistry.RegistryUrl(), currentRegistry.GetAuthConfig())

		// If ECR cache is configured, authenticate to the cache registry separately
		// This works even when the image registry is not ECR (e.g., GitLab + ECR cache)
		if cfg.Cache.ECR.Configured() {
			cacheRegistry := cfg.Cache.ECR.AsRegistry()
			if cacheRegistry.Configured() {
				log.Debugf("Authenticating against ECR cache registry\n")
				if err := cacheRegistry.Login(client); err != nil {
					log.Warnf("Failed to authenticate to ECR cache registry: %v\n", err)
				} else {
					cacheHost := extractHost(cfg.Cache.ECR.Url)
					log.Debugf("Adding cache registry credentials for <green>%s</green>\n", cacheHost)
					authenticator.AddCredentials(cacheHost, cacheRegistry.GetAuthConfig())
				}
			}
		}
	}

	var content []byte
	if buildVars.isDockerfileFromStdin() {
		log.Infof("<greed>reading Dockerfile content from stdin</green>\n")
		content, err = io.ReadAll(buildVars.StdIn)
	} else {
		content, err = os.ReadFile(filepath.Join(dir, buildVars.Dockerfile))
	}
	if err != nil {
		log.Error(fmt.Sprintf("<red>%s</red>", err.Error()))
		return err
	}
	dockerFile, err := os.Create(filepath.Join(dir, "build-tools-dockerfile"))
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(dockerFile.Name()) }()
	if _, err := dockerFile.Write(content); err != nil {
		return err
	}
	if err := dockerFile.Close(); err != nil {
		return err
	}
	buildVars.Dockerfile, err = filepath.Rel(dir, dockerFile.Name())
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(content))) == 0 {
		return fmt.Errorf("<red>the Dockerfile cannot be empty</red>")
	}
	stages := docker.FindStages(string(content))
	if !ci.IsValid(currentCI) {
		return fmt.Errorf("commit and/or branch information is <red>missing</red> (perhaps you're not in a Git repository or forgot to set environment variables?)")
	}

	commit := currentCI.Commit()
	branch := currentCI.BranchReplaceSlash()
	log.Debugf("Using build variables commit <green>%s</green> on branch <green>%s</green>\n", commit, branch)
	var tags []string
	branchTag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), branch)
	latestTag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), "latest")
	tags = append(tags, []string{
		docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), commit),
		branchTag,
	}...)
	if currentCI.Branch() == "master" || currentCI.Branch() == "main" {
		tags = append(tags, latestTag)
	}

	caches := []string{branchTag, latestTag}

	buildArgs := map[string]*string{
		"BUILDKIT_INLINE_CACHE": aws.String("1"),
		"CI_COMMIT":             &commit,
		"CI_BRANCH":             &branch,
	}
	for _, arg := range buildVars.BuildArgs {
		split := strings.Split(arg, "=")
		key := split[0]
		value := strings.Join(split[1:], "=")
		if len(split) > 1 && len(value) > 0 {
			buildArgs[key] = &value
		} else {
			if env, exists := os.LookupEnv(key); exists {
				buildArgs[key] = &env
			} else {
				log.Debugf("ignoring build-arg %s\n", key)
			}
		}
	}

	ecrCache := cfg.Cache.ECR
	for _, stage := range stages {
		tag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), stage)
		caches = append([]string{tag}, caches...)
		err := buildStage(client, dir, buildVars, buildArgs, []string{tag}, caches, stage, ecrCache, authenticator)
		if err != nil {
			return err
		}
	}

	if buildVars.isMultiPlatform() {
		return buildMultiPlatform(client, dir, buildVars, buildArgs, tags, caches, "", ecrCache, authenticator)
	}
	return buildStage(client, dir, buildVars, buildArgs, tags, caches, "", ecrCache, authenticator)
}

func buildStage(dkrClient docker.Client, dir string, buildVars Args, buildArgs map[string]*string, tags []string, caches []string, stage string, ecrCache *config.ECRCache, authenticator docker.Authenticator) error {
	// If BUILDKIT_HOST is set, use buildkit client directly (pushes to registry)
	if os.Getenv("BUILDKIT_HOST") != "" {
		return buildMultiPlatform(dkrClient, dir, buildVars, buildArgs, tags, caches, stage, ecrCache, authenticator)
	}

	// Otherwise use Docker API (loads to local daemon)
	s := setupSession(dir)
	if authenticator != nil {
		s.Allow(authenticator)
	}
	fs, err := fsutil.NewFS(dir)
	if err != nil {
		return err
	}
	s.Allow(filesync.NewFSSyncProvider(filesync.StaticDirSource{
		"context":    fs,
		"dockerfile": fs,
	}))
	s.Allow(filesync.NewFSSyncTarget(filesync.WithFSSyncDir(0, "exported")))

	eg, ctx := errgroup.WithContext(context.Background())
	dialSession := func(ctx context.Context, proto string, meta map[string][]string) (net.Conn, error) {
		return dkrClient.DialHijack(ctx, "/session", proto, meta)
	}
	eg.Go(func() error {
		return s.Run(context.TODO(), dialSession)
	})
	eg.Go(func() error {
		defer func() { // make sure the Status ends cleanly on build errors
			_ = s.Close()
		}()
		var outputs []dockerbuild.ImageBuildOutput
		if strings.HasPrefix(stage, "export") {
			outputs = append(outputs, dockerbuild.ImageBuildOutput{
				Type:  "local",
				Attrs: map[string]string{},
			})
		}
		sessionID := s.ID()
		return doBuild(ctx, dkrClient, eg, buildVars.dockerfileName(), buildArgs, tags, caches, stage, !buildVars.NoPull, sessionID, outputs, buildVars.Platform)
	})
	return eg.Wait()
}

// buildFrontendAttrs creates the frontend attributes map for buildkit.
// It includes the dockerfile name, platform, target stage, and any build arguments.
func buildFrontendAttrs(dockerfile, platform, target string, buildArgs map[string]*string) map[string]string {
	attrs := map[string]string{
		"filename": dockerfile,
		"platform": platform,
	}

	if target != "" {
		attrs["target"] = target
	}

	for k, v := range buildArgs {
		if v != nil {
			attrs["build-arg:"+k] = *v
		}
	}

	return attrs
}

// buildExportEntry creates the export configuration for pushing images to a registry.
// All tags are joined into a single entry with comma-separated names.
func buildExportEntry(tags []string) client.ExportEntry {
	return client.ExportEntry{
		Type: client.ExporterImage,
		Attrs: map[string]string{
			"name": strings.Join(tags, ","),
			"push": "true",
		},
	}
}

// buildCacheImports creates cache import entries for registry-based caching.
// ECR cache is added first (if configured) so it's checked before other caches.
func buildCacheImports(caches []string, ecrCache *config.ECRCache) []client.CacheOptionsEntry {
	var imports []client.CacheOptionsEntry
	// Add ECR cache import first if configured (checked before other caches)
	if ecrCache.Configured() {
		imports = append(imports, client.CacheOptionsEntry{
			Type: "registry",
			Attrs: map[string]string{
				"ref": ecrCache.CacheRef(),
			},
		})
	}
	for _, cache := range caches {
		imports = append(imports, client.CacheOptionsEntry{
			Type: "registry",
			Attrs: map[string]string{
				"ref": cache,
			},
		})
	}
	return imports
}

// buildCacheExports creates cache export entries for ECR registry-based caching.
// ECR requires special settings: image-manifest=true and oci-mediatypes=true.
// See: https://aws.amazon.com/blogs/containers/announcing-remote-cache-support-in-amazon-ecr-for-buildkit-clients/
func buildCacheExports(ecrCache *config.ECRCache) []client.CacheOptionsEntry {
	if !ecrCache.Configured() {
		return nil
	}
	return []client.CacheOptionsEntry{
		{
			Type: "registry",
			Attrs: map[string]string{
				"ref":            ecrCache.CacheRef(),
				"mode":           "max",
				"image-manifest": "true",
				"oci-mediatypes": "true",
			},
		},
	}
}

// hasContainerdSnapshotter checks if any worker has containerd snapshotter support.
// This is required for multi-platform builds via Docker's embedded buildkit.
func hasContainerdSnapshotter(workers []*client.WorkerInfo) bool {
	for _, w := range workers {
		for key, value := range w.Labels {
			if strings.Contains(key, "containerd") || strings.Contains(value, "containerd") {
				return true
			}
		}
	}
	return false
}

// buildMultiPlatform builds images for multiple platforms using buildkit client.
// This is required because Docker's /build API only supports single platform builds.
//
// Connection priority:
// 1. If BUILDKIT_HOST is set, connect directly to that buildkit instance
// 2. Otherwise, connect via Docker's /grpc endpoint (requires containerd snapshotter)
//
// For Docker's embedded buildkit, containerd snapshotter must be enabled.
// See: https://docs.docker.com/storage/containerd/
// Enable it by adding to /etc/docker/daemon.json:
//
//	{ "features": { "containerd-snapshotter": true } }
func buildMultiPlatform(dkrClient docker.Client, dir string, buildVars Args, buildArgs map[string]*string, tags []string, caches []string, target string, ecrCache *config.ECRCache, authenticator docker.Authenticator) error {
	return buildMultiPlatformWithFactory(dkrClient, dir, buildVars, buildArgs, tags, caches, target, ecrCache, authenticator, defaultBuildkitClientFactory)
}

// buildMultiPlatformWithFactory is the internal implementation that accepts a BuildkitClientFactory for testing.
func buildMultiPlatformWithFactory(dkrClient docker.Client, dir string, buildVars Args, buildArgs map[string]*string, tags []string, caches []string, target string, ecrCache *config.ECRCache, authenticator docker.Authenticator, clientFactory BuildkitClientFactory) error {
	fs, err := fsutil.NewFS(dir)
	if err != nil {
		return err
	}

	eg, ctx := errgroup.WithContext(context.Background())

	var bkClient BuildkitClient

	// Check if BUILDKIT_HOST is set - if so, connect directly to buildkit
	buildkitHost := os.Getenv("BUILDKIT_HOST")
	if buildkitHost != "" {
		log.Infof("Connecting to buildkit at <green>%s</green>\n", buildkitHost)
		bkClient, err = clientFactory(ctx, buildkitHost)
		if err != nil {
			return fmt.Errorf("failed to connect to buildkit at %s: %w", buildkitHost, err)
		}
	} else {
		// Connect to buildkit via Docker's grpc endpoint (like buildx does)
		log.Debug("Connecting to buildkit via Docker daemon")
		dialContext := func(ctx context.Context, _ string) (net.Conn, error) {
			return dkrClient.DialHijack(ctx, "/grpc", "h2c", nil)
		}
		dialSession := func(ctx context.Context, proto string, meta map[string][]string) (net.Conn, error) {
			return dkrClient.DialHijack(ctx, "/session", proto, meta)
		}

		bkClient, err = clientFactory(ctx, "",
			client.WithContextDialer(dialContext),
			client.WithSessionDialer(dialSession),
		)
		if err != nil {
			return fmt.Errorf("failed to create buildkit client: %w", err)
		}

		// Check if the image exporter is available (requires containerd snapshotter)
		// by querying worker info - only needed when using Docker's embedded buildkit
		workers, err := bkClient.ListWorkers(ctx)
		if err != nil {
			return fmt.Errorf("failed to list buildkit workers: %w", err)
		}

		if !hasContainerdSnapshotter(workers) {
			log.Warn("Docker may not have containerd snapshotter enabled. Multi-platform builds require it.")
			log.Warn("Alternatively, set BUILDKIT_HOST to connect to a standalone buildkit instance.")
			log.Warn("Enable containerd snapshotter by adding to /etc/docker/daemon.json: {\"features\": {\"containerd-snapshotter\": true}}")
		}
	}
	defer func() { _ = bkClient.Close() }()

	frontendAttrs := buildFrontendAttrs(buildVars.dockerfileName(), buildVars.Platform, target, buildArgs)
	cacheImports := buildCacheImports(caches, ecrCache)
	cacheExports := buildCacheExports(ecrCache)

	// Determine export type: local for "export" stages, image for regular builds
	var exports []client.ExportEntry
	isExportStage := strings.HasPrefix(target, "export")
	if isExportStage {
		// Export to local "exported" directory
		exportDir := filepath.Join(dir, "exported")
		if err := os.MkdirAll(exportDir, 0o755); err != nil {
			return fmt.Errorf("failed to create export directory: %w", err)
		}
		exports = []client.ExportEntry{{
			Type:      client.ExporterLocal,
			OutputDir: exportDir,
		}}
		log.Infof("Exporting build artifacts to <green>%s</green>\n", exportDir)
	} else {
		exports = []client.ExportEntry{buildExportEntry(tags)}
	}

	if ecrCache.Configured() {
		log.Infof("Using ECR cache at <green>%s</green>\n", ecrCache.CacheRef())
	}

	// Session attachables for authentication and file sync
	var sessionAttachables []session.Attachable
	if authenticator != nil {
		sessionAttachables = append(sessionAttachables, authenticator)
		log.Debugf("Added authenticator for multi-platform build\n")
	} else if !isExportStage {
		log.Warnf("No authenticator provided for multi-platform build - registry push may fail\n")
	}

	solveOpt := client.SolveOpt{
		Frontend:      "dockerfile.v0",
		FrontendAttrs: frontendAttrs,
		LocalMounts: map[string]fsutil.FS{
			"context":    fs,
			"dockerfile": fs,
		},
		Session:      sessionAttachables,
		Exports:      exports,
		CacheImports: cacheImports,
		CacheExports: cacheExports,
	}

	// Create status channel for progress
	statusChan := make(chan *client.SolveStatus)
	displayStatus(os.Stdout, statusChan, eg)

	_, err = bkClient.Solve(ctx, nil, solveOpt, statusChan)
	if err != nil {
		if strings.Contains(err.Error(), "exporter") && strings.Contains(err.Error(), "not be found") {
			log.Error("This error typically means Docker's containerd snapshotter is not enabled")
			log.Error("Enable it by adding to /etc/docker/daemon.json: {\"features\": {\"containerd-snapshotter\": true}}")
			log.Error("Then restart Docker")
			return fmt.Errorf("multi-platform build failed: %w", err)
		}
		return fmt.Errorf("multi-platform build failed: %w", err)
	}

	log.Info("Build successful")
	return nil
}

func doBuild(ctx context.Context, dkrClient docker.Client, eg *errgroup.Group, dockerfile string, args map[string]*string, tags, caches []string, target string, pullParent bool, sessionID string, outputs []dockerbuild.ImageBuildOutput, platform string) (finalErr error) {
	buildID := stringid.GenerateRandomID()
	options := dockerbuild.ImageBuildOptions{
		BuildArgs:     args,
		BuildID:       buildID,
		CacheFrom:     caches,
		Dockerfile:    dockerfile,
		Outputs:       outputs,
		PullParent:    pullParent,
		MemorySwap:    -1,
		RemoteContext: "client-session",
		Remove:        true,
		SessionID:     sessionID,
		ShmSize:       256 * 1024 * 1024,
		Tags:          tags,
		Target:        target,
		Version:       dockerbuild.BuilderBuildKit,
		Platform:      platform,
	}
	logVerbose(options)
	var response dockerbuild.ImageBuildResponse
	var err error
	response, err = dkrClient.ImageBuild(context.Background(), nil, options)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()

	done := make(chan struct{})
	defer close(done)
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return dkrClient.BuildCancel(context.TODO(), buildID)
		case <-done:
		}
		return nil
	})

	tracer := newTracer()

	displayStatus(os.Stderr, tracer.displayCh, eg)
	defer close(tracer.displayCh)

	buf := &bytes.Buffer{}
	imageID := ""
	writeAux := func(msg jsonmessage.JSONMessage) {
		if msg.ID == "moby.image.id" {
			var result dockerbuild.Result
			if err := json.Unmarshal(*msg.Aux, &result); err != nil {
				log.Errorf("failed to parse aux message: %v", err)
			}
			imageID = result.ID
			return
		}
		tracer.write(msg)
	}

	err = jsonmessage.DisplayJSONMessagesStream(response.Body, buf, os.Stdout.Fd(), true, writeAux)
	if err != nil {
		var jerr *jsonmessage.JSONError
		if errors.As(err, &jerr) {
			// If no error code is set, default to 1
			if jerr.Code == 0 {
				jerr.Code = 1
			}
			return fmt.Errorf("code: %d, status: %s", jerr.Code, jerr.Message)
		}
	}

	imageID = buf.String()
	log.Info(imageID)

	return nil
}

func displayStatus(out *os.File, displayCh chan *client.SolveStatus, eg *errgroup.Group) {
	// not using shared context to not disrupt display but let it finish reporting errors
	display, err := progressui.NewDisplay(out, progressui.AutoMode)
	if err != nil {
		eg.Go(func() error {
			return err
		})
	}
	eg.Go(func() error {
		_, err := display.UpdateFrom(context.TODO(), displayCh)
		return err
	})
}

func logVerbose(options dockerbuild.ImageBuildOptions) {
	loggableOptions := options
	loggableOptions.AuthConfigs = nil
	loggableOptions.BuildID = ""
	loggableOptions.SessionID = ""
	marshal, _ := yaml.Marshal(loggableOptions)
	log.Debugf("performing docker build with options (auths removed):\n%s\n", marshal)
}

func getBuildSharedKey(dir string) string {
	// build session id hash of build dir with node based randomness
	s := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", tryNodeIdentifier(), dir)))
	return hex.EncodeToString(s[:])
}

func tryNodeIdentifier() string {
	out := filepath.Join(os.TempDir(), ".buildtools") // return config dir as default on permission error
	if err := os.MkdirAll(out, 0o700); err == nil {
		sessionFile := filepath.Join(out, ".buildNodeID")
		if _, err := os.Lstat(sessionFile); err != nil {
			if os.IsNotExist(err) { // create a new file with stored randomness
				b := make([]byte, 32)
				if _, err := rand.Read(b); err != nil {
					return out
				}
				if err := os.WriteFile(sessionFile, []byte(hex.EncodeToString(b)), 0o600); err != nil {
					return out
				}
			}
		}

		dt, err := os.ReadFile(sessionFile)
		if err == nil {
			return string(dt)
		}
	}
	return out
}

type tracer struct {
	displayCh chan *client.SolveStatus
}

func newTracer() *tracer {
	return &tracer{
		displayCh: make(chan *client.SolveStatus),
	}
}

func (t *tracer) write(msg jsonmessage.JSONMessage) {
	var resp controlapi.StatusResponse

	if msg.ID != "moby.buildkit.trace" {
		return
	}

	var dt []byte
	// ignoring all messages that are not understood
	if err := json.Unmarshal(*msg.Aux, &dt); err != nil {
		return
	}
	if err := (&resp).UnmarshalVT(dt); err != nil {
		return
	}

	s := client.SolveStatus{}
	for _, v := range resp.Vertexes {
		inputs := make([]digest.Digest, len(v.Inputs))
		for i, input := range v.Inputs {
			inputs[i] = digest.Digest(input)
		}
		var started *time.Time
		var completed *time.Time
		if v.Started != nil {
			t := v.Started.AsTime()
			started = &t
		}
		if v.Completed != nil {
			t := v.Completed.AsTime()
			completed = &t
		}
		s.Vertexes = append(s.Vertexes, &client.Vertex{
			Digest:    digest.Digest(v.Digest),
			Inputs:    inputs,
			Name:      v.Name,
			Started:   started,
			Completed: completed,
			Error:     v.Error,
			Cached:    v.Cached,
		})
	}
	for _, v := range resp.Statuses {
		var started *time.Time
		var completed *time.Time
		if v.Started != nil {
			t := v.Started.AsTime()
			started = &t
		}
		if v.Completed != nil {
			t := v.Completed.AsTime()
			completed = &t
		}
		s.Statuses = append(s.Statuses, &client.VertexStatus{
			ID:        v.ID,
			Vertex:    digest.Digest(v.Vertex),
			Name:      v.Name,
			Total:     v.Total,
			Current:   v.Current,
			Timestamp: v.Timestamp.AsTime(),
			Started:   started,
			Completed: completed,
		})
	}
	for _, v := range resp.Logs {
		s.Logs = append(s.Logs, &client.VertexLog{
			Vertex:    digest.Digest(v.Vertex),
			Stream:    int(v.Stream),
			Data:      v.Msg,
			Timestamp: v.Timestamp.AsTime(),
		})
	}

	t.displayCh <- &s
}
