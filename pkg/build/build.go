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
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stringid"
	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/buildkit/client"
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
	Dockerfile string   `name:"file" short:"f" help:"name of the Dockerfile to use." default:"Dockerfile"`
	BuildArgs  []string `name:"build-arg" type:"list" help:"additional docker build-args to use, see https://docs.docker.com/engine/reference/commandline/build/ for more information."`
	NoLogin    bool     `help:"disable login to docker registry" default:"false" `
	NoPull     bool     `help:"disable pulling latest from docker registry" default:"false"`
	Platform   string   `help:"specify target platform to build" default:""`
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
		log.Infof("building for platform <green>%s</green>\n", buildVars.Platform)
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
	}

	content, err := os.ReadFile(filepath.Join(dir, buildVars.Dockerfile))
	if err != nil {
		log.Error(fmt.Sprintf("<red>%s</red>", err.Error()))
		return err
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

	for _, stage := range stages {
		tag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), stage)
		caches = append([]string{tag}, caches...)
		err := buildStage(client, dir, buildVars, buildArgs, []string{tag}, caches, stage, authenticator)
		if err != nil {
			return err
		}
	}

	return buildStage(client, dir, buildVars, buildArgs, tags, caches, "", authenticator)
}

func buildStage(client docker.Client, dir string, buildVars Args, buildArgs map[string]*string, tags []string, caches []string, stage string, authenticator docker.Authenticator) error {
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
		return client.DialHijack(ctx, "/session", proto, meta)
	}
	eg.Go(func() error {
		return s.Run(context.TODO(), dialSession)
	})
	eg.Go(func() error {
		defer func() { // make sure the Status ends cleanly on build errors
			_ = s.Close()
		}()
		var outputs []types.ImageBuildOutput
		if strings.HasPrefix(stage, "export") {
			outputs = append(outputs, types.ImageBuildOutput{
				Type:  "local",
				Attrs: map[string]string{},
			})
		}
		sessionID := s.ID()
		return doBuild(ctx, client, eg, buildVars.Dockerfile, buildArgs, tags, caches, stage, !buildVars.NoPull, sessionID, outputs, buildVars.Platform)
	})
	return eg.Wait()
}

func doBuild(ctx context.Context, dkrClient docker.Client, eg *errgroup.Group, dockerfile string, args map[string]*string, tags, caches []string, target string, pullParent bool, sessionID string, outputs []types.ImageBuildOutput, platform string) (finalErr error) {
	buildID := stringid.GenerateRandomID()
	options := types.ImageBuildOptions{
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
		Version:       types.BuilderBuildKit,
		Platform:      platform,
	}
	logVerbose(options)
	var response types.ImageBuildResponse
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
			var result types.BuildResult
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
		if jerr, ok := err.(*jsonmessage.JSONError); ok {
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

func logVerbose(options types.ImageBuildOptions) {
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
	if err := os.MkdirAll(out, 0700); err == nil {
		sessionFile := filepath.Join(out, ".buildNodeID")
		if _, err := os.Lstat(sessionFile); err != nil {
			if os.IsNotExist(err) { // create a new file with stored randomness
				b := make([]byte, 32)
				if _, err := rand.Read(b); err != nil {
					return out
				}
				if err := os.WriteFile(sessionFile, []byte(hex.EncodeToString(b)), 0600); err != nil {
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
