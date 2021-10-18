package build

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/apex/log"
	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/ci"
	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/docker"
	"github.com/buildtool/build-tools/pkg/tar"
	"github.com/containerd/console"
	"github.com/docker/docker/api/types"
	dkr "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stringid"
	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/session/filesync"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/moby/buildkit/util/progress/progresswriter"
	fsutiltypes "github.com/tonistiigi/fsutil/types"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

type Args struct {
	args.Globals
	Dockerfile string   `name:"file" short:"f" help:"name of the Dockerfile to use." default:"Dockerfile"`
	BuildArgs  []string `name:"build-arg" type:"list" help:"additional docker build-args to use, see https://docs.docker.com/engine/reference/commandline/build/ for more information."`
	NoLogin    bool     `help:"disable login to docker registry" default:"false" `
	NoPull     bool     `help:"disable pulling latest from docker registry" default:"false"`
}

func DoBuild(dir string, buildArgs Args) error {
	dkrClient, err := dockerClient()
	if err != nil {
		return err
	}
	buildContext, err := createBuildContext(dir, buildArgs.Dockerfile)
	if err != nil {
		return err
	}
	return build(dkrClient, dir, buildContext, buildArgs)
}

var dockerClient = func() (docker.Client, error) {
	return dkr.NewClientWithOpts(dkr.FromEnv)
}

func createBuildContext(dir, dockerfile string) (io.ReadCloser, error) {
	ignored, err := docker.ParseDockerignore(dir, dockerfile)
	if err != nil {
		return nil, err
	}
	return archive.TarWithOptions(dir, &archive.TarOptions{ExcludePatterns: ignored})
}

func setupSession(dir string) (*session.Session, session.Attachable) {
	s, err := session.NewSession(context.Background(), filepath.Base(dir), getBuildSharedKey(dir))
	if err != nil {
		panic("session.NewSession changed behaviour and returned an error. Create an issue at https://github.com/buildtool/build-tools/issues/new")
	}
	if s == nil {
		panic("session.NewSession changed behaviour and did not return a session. Create an issue at https://github.com/buildtool/build-tools/issues/new")
	}

	dockerAuthProvider := authprovider.NewDockerAuthProvider(os.Stderr)
	s.Allow(dockerAuthProvider)

	s.Allow(filesync.NewFSSyncProvider([]filesync.SyncedDir{
		{
			Name: "context",
			Dir:  dir,
			Map:  resetUIDAndGID,
		},
		{
			Name: "dockerfile",
			Dir:  dir,
		},
	}))

	return s, dockerAuthProvider
}

func build(client docker.Client, dir string, buildContext io.ReadCloser, buildVars Args) error {
	cfg, err := config.Load(dir)
	if err != nil {
		return err
	}
	currentCI := cfg.CurrentCI()
	log.Debugf("Using CI <green>%s</green>\n", currentCI.Name())

	currentRegistry := cfg.CurrentRegistry()
	log.Debugf("Using registry <green>%s</green>\n", currentRegistry.Name())
	authConfigs := make(map[string]types.AuthConfig)
	if buildVars.NoLogin {
		log.Debugf("Login <yellow>disabled</yellow>\n")
	} else {
		log.Debugf("Authenticating against registry <green>%s</green>\n", currentRegistry.Name())
		if err := currentRegistry.Login(client); err != nil {
			return err
		}
		authConfigs[currentRegistry.RegistryUrl()] = currentRegistry.GetAuthConfig()
	}

	var buf bytes.Buffer
	tee := io.TeeReader(buildContext, &buf)
	stages, err := findStages(tee, buildVars.Dockerfile)
	if err != nil {
		return err
	}
	if !ci.IsValid(currentCI) {
		return fmt.Errorf("commit and/or branch information is <red>missing</red> (perhaps you're not in a Git repository or forgot to set environment variables?)")
	}

	commit := currentCI.Commit()
	branch := currentCI.BranchReplaceSlash()
	log.Debugf("Using build variables commit <green>%s</green> on branch <green>%s</green>\n", commit, branch)
	var caches []string

	buildArgs := map[string]*string{
		"CI_COMMIT": &commit,
		"CI_BRANCH": &branch,
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
		err := buildStage(client, dir, buildVars, buildArgs, []string{tag}, caches, stage, authConfigs)
		if err != nil {
			return err
		}
	}

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

	caches = append([]string{branchTag, latestTag}, caches...)
	return buildStage(client, dir, buildVars, buildArgs, tags, caches, "", authConfigs)
}

func buildStage(client docker.Client, dir string, buildVars Args, buildArgs map[string]*string, tags []string, caches []string, stage string, authConfigs map[string]types.AuthConfig) error {
	s, dockerAuthProvider := setupSession(dir)
	eg, ctx := errgroup.WithContext(context.Background())
	dialSession := func(ctx context.Context, proto string, meta map[string][]string) (net.Conn, error) {
		return client.DialHijack(ctx, "/session", proto, meta)
	}
	m := &sync.Mutex{}
	eg.Go(func() error {
		m.Lock()
		defer m.Unlock()
		return s.Run(context.TODO(), dialSession)
	})
	eg.Go(func() error {
		defer func() { // make sure the Status ends cleanly on build errors
			m.Lock()
			defer m.Unlock()
			err := s.Close()
			if err != nil {
				log.Errorf("error closing session: %s\n", err)
			}
		}()
		var outputs []types.ImageBuildOutput
		if strings.HasPrefix(stage, "export") {
			outputs = append(outputs, types.ImageBuildOutput{
				Type:  "local",
				Attrs: map[string]string{},
			})
			func() {
				m.Lock()
				defer m.Unlock()
				s.Allow(filesync.NewFSSyncTargetDir("exported"))
			}()
		}
		sessionID := s.ID()
		return doBuild(ctx, client, eg, buildVars.Dockerfile, buildArgs, tags, caches, stage, authConfigs, !buildVars.NoPull, sessionID, dockerAuthProvider, outputs)
	})
	return eg.Wait()
}

func doBuild(ctx context.Context, dkrClient docker.Client, eg *errgroup.Group, dockerfile string, args map[string]*string, tags, caches []string, target string, authConfigs map[string]types.AuthConfig, pullParent bool, sessionID string, at session.Attachable, outputs []types.ImageBuildOutput) (finalErr error) {
	buildID := stringid.GenerateRandomID()
	options := types.ImageBuildOptions{
		AuthConfigs:   authConfigs,
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

	displayStatus(os.Stderr, tracer.displayCh, eg, at)
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

func displayStatus(out *os.File, displayCh chan *client.SolveStatus, eg *errgroup.Group, at session.Attachable) {
	var c console.Console
	// TODO: Handle tty output in non-tty environment.
	if cons, err := console.ConsoleFromFile(out); err == nil {
		c = cons
	}
	// not using shared context to not disrupt display but let it finish reporting errors
	eg.Go(func() error {
		return progressui.DisplaySolveStatus(context.TODO(), "", c, out, displayCh)
	})
	if s, ok := at.(interface {
		SetLogger(progresswriter.Logger)
	}); ok {
		s.SetLogger(func(s *client.SolveStatus) {
			displayCh <- s
		})
	}
}

func logVerbose(options types.ImageBuildOptions) {
	loggableOptions := options
	loggableOptions.AuthConfigs = nil
	loggableOptions.BuildID = ""
	loggableOptions.SessionID = ""
	marshal, _ := yaml.Marshal(loggableOptions)
	log.Debugf("performing docker build with options (auths removed):\n%s\n", marshal)
}

func findStages(buildContext io.Reader, dockerfile string) ([]string, error) {
	content, err := tar.ExtractFileContent(buildContext, dockerfile)
	if err != nil {
		return nil, err
	}
	stages := docker.FindStages(content)

	return stages, nil
}

func getBuildSharedKey(dir string) string {
	// build session is hash of build dir with node based randomness
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
				if err := ioutil.WriteFile(sessionFile, []byte(hex.EncodeToString(b)), 0600); err != nil {
					return out
				}
			}
		}

		dt, err := ioutil.ReadFile(sessionFile)
		if err == nil {
			return string(dt)
		}
	}
	return out
}

func resetUIDAndGID(_ string, s *fsutiltypes.Stat) bool {
	s.Uid = 0
	s.Gid = 0
	return true
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
	if err := (&resp).Unmarshal(dt); err != nil {
		return
	}

	s := client.SolveStatus{}
	for _, v := range resp.Vertexes {
		s.Vertexes = append(s.Vertexes, &client.Vertex{
			Digest:    v.Digest,
			Inputs:    v.Inputs,
			Name:      v.Name,
			Started:   v.Started,
			Completed: v.Completed,
			Error:     v.Error,
			Cached:    v.Cached,
		})
	}
	for _, v := range resp.Statuses {
		s.Statuses = append(s.Statuses, &client.VertexStatus{
			ID:        v.ID,
			Vertex:    v.Vertex,
			Name:      v.Name,
			Total:     v.Total,
			Current:   v.Current,
			Timestamp: v.Timestamp,
			Started:   v.Started,
			Completed: v.Completed,
		})
	}
	for _, v := range resp.Logs {
		s.Logs = append(s.Logs, &client.VertexLog{
			Vertex:    v.Vertex,
			Stream:    int(v.Stream),
			Data:      v.Msg,
			Timestamp: v.Timestamp,
		})
	}

	t.displayCh <- &s
}
