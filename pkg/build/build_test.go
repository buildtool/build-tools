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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apex/log"
	dockerbuild "github.com/docker/docker/api/types/build"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/client/llb"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/stretchr/testify/assert"

	"github.com/buildtool/build-tools/pkg"
	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/docker"
)

// MockBuildkitClient is a mock implementation of BuildkitClient for testing.
type MockBuildkitClient struct {
	SolveFunc       func(ctx context.Context, def *llb.Definition, opt client.SolveOpt, statusChan chan *client.SolveStatus) (*client.SolveResponse, error)
	ListWorkersFunc func(ctx context.Context, opts ...client.ListWorkersOption) ([]*client.WorkerInfo, error)
	CloseFunc       func() error
}

func (m *MockBuildkitClient) Solve(ctx context.Context, def *llb.Definition, opt client.SolveOpt, statusChan chan *client.SolveStatus) (*client.SolveResponse, error) {
	if m.SolveFunc != nil {
		return m.SolveFunc(ctx, def, opt, statusChan)
	}
	close(statusChan)
	return &client.SolveResponse{}, nil
}

func (m *MockBuildkitClient) ListWorkers(ctx context.Context, opts ...client.ListWorkersOption) ([]*client.WorkerInfo, error) {
	if m.ListWorkersFunc != nil {
		return m.ListWorkersFunc(ctx, opts...)
	}
	return []*client.WorkerInfo{}, nil
}

func (m *MockBuildkitClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

var name string

func TestMain(m *testing.M) {
	buildToolsTempDir, osTempDir := setup()
	setupSession = func(dir string) Session {
		return &MockSession{}
	}
	code := m.Run()
	teardown(buildToolsTempDir, osTempDir)
	os.Exit(code)
}

func setup() (string, string) {
	name, _ = os.MkdirTemp(os.TempDir(), "build-tools")
	temp, _ := os.MkdirTemp(os.TempDir(), "build-tools-temp")
	os.Clearenv()
	_ = os.Setenv("TMPDIR", temp)
	_ = os.Setenv("TEMP", temp)

	return name, temp
}

func teardown(buildToolsTempDir, osTempDir string) {
	_ = os.RemoveAll(buildToolsTempDir)
	_ = os.RemoveAll(osTempDir)
}

func TestBuild_BrokenConfig(t *testing.T) {
	yaml := `ci: [] `
	file := filepath.Join(name, ".buildtools.yaml")
	_ = os.WriteFile(file, []byte(yaml), 0o777)
	defer func() { _ = os.Remove(file) }()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")
	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	absPath, _ := filepath.Abs(filepath.Join(name, ".buildtools.yaml"))
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig")
	logMock.Check(t, []string{fmt.Sprintf("debug: Parsing config from file: <green>'%s'</green>\n", absPath)})
}

func TestBuild_NoRegistry(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	tmpDockerClient := dockerClient
	dockerClient = func() (client docker.Client, e error) {
		return &docker.MockDocker{}, nil
	}
	defer func() { dockerClient = tmpDockerClient }()

	_, err := DoBuild(name, Args{Dockerfile: "Dockerfile"})
	assert.NoError(t, err)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>No docker registry</green>\n",
		"debug: Authenticating against registry <green>No docker registry</green>\n",
		"debug: Authentication <yellow>not supported</yellow> for registry <green>No docker registry</green>\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>feature1</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - noregistry/reponame:abc123\n    - noregistry/reponame:feature1\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - noregistry/reponame:feature1\n    - noregistry/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
	})
}

func TestBuild_LoginError(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")
	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.EqualError(t, err, "invalid username/password")
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"error: Unable to login\n",
	})
}

func TestBuild_NoCI(t *testing.T) {
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{BuildError: []error{fmt.Errorf("build error")}}

	_ = write(name, "Dockerfile", "FROM scratch")
	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.EqualError(t, err, "commit and/or branch information is <red>missing</red> (perhaps you're not in a Git repository or forgot to set environment variables?)")
	logMock.Check(t, []string{
		"debug: Using CI <green>none</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
	})
}

func TestBuild_BuildError(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{BuildError: []error{fmt.Errorf("build error")}}
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")
	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.EqualError(t, err, "build error")
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>feature1</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
	})
}

func TestBuild_BuildResponseError(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{ResponseError: fmt.Errorf("build error")}
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")
	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.EqualError(t, err, "code: 123, status: build error")
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>feature1</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
	})
}

func TestBuild_ErrorOutput(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{BrokenOutput: true}
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")
	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.EqualError(t, err, "code: 1, status: some message")
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>feature1</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
	})
}

func TestBuild_ValidOutput(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	f, err := os.Open("testdata/build_body.txt")
	assert.NoError(t, err)
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{ResponseBody: bufio.NewReader(f)}
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")
	_, err = build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>feature1</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: 1: msg 1\n2: msg 2\n",
	})
}

func TestBuild_BrokenBuildResult(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{ResponseBody: strings.NewReader(`{"id":"moby.image.id","aux":{"id":123}}`)}
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")
	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>feature1</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"error: failed to parse aux message: json: cannot unmarshal number into Go struct field Result.ID of type string",
		"info: ",
	})
}

func TestBuild_WithBuildArgs(t *testing.T) {
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("CI_COMMIT_SHA", "sha")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	client := &docker.MockDocker{}
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  []string{"buildargs1=1", "buildargs2=2"},
		NoLogin:    false,
		NoPull:     false,
	})
	assert.NoError(t, err)

	assert.Equal(t, 5, len(client.BuildOptions[0].BuildArgs))
	assert.Equal(t, "1", *client.BuildOptions[0].BuildArgs["buildargs1"])
	assert.Equal(t, "2", *client.BuildOptions[0].BuildArgs["buildargs2"])
}

func TestBuild_WithStrangeBuildArg(t *testing.T) {
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("CI_COMMIT_SHA", "sha")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("buildargs4", "env-value")()
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  []string{"buildargs1=1=1", "buildargs2", "buildargs3=", "buildargs4"},
		NoLogin:    false,
		NoPull:     false,
	})
	assert.NoError(t, err)

	assert.Equal(t, 5, len(client.BuildOptions[0].BuildArgs))
	assert.Equal(t, "1=1", *client.BuildOptions[0].BuildArgs["buildargs1"])
	assert.Equal(t, "env-value", *client.BuildOptions[0].BuildArgs["buildargs4"])
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>sha</green> on branch <green>master</green>\n",
		"debug: ignoring build-arg buildargs2\n",
		"debug: ignoring build-arg buildargs3\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:sha\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: sha\n    buildargs1: 1=1\n    buildargs4: env-value\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
	})
}

func TestBuild_WithPlatform(t *testing.T) {
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("CI_COMMIT_SHA", "sha")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		NoLogin:    false,
		NoPull:     false,
		Platform:   "linux/amd64",
	})
	assert.NoError(t, err)

	logMock.Check(t, []string{
		"info: building for platform <green>linux/amd64</green>\n",
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>sha</green> on branch <green>master</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:sha\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: sha\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: linux/amd64\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
	})
}

// TestBuild_WithMultiPlatform is an integration test that requires a running Docker daemon with buildkit.
// Multi-platform builds use the buildkit client directly via Docker's /grpc endpoint,
// which cannot be easily mocked. This test verifies that multi-platform detection works correctly.
// For actual multi-platform build testing, use integration tests with a real Docker daemon.
func TestBuild_WithMultiPlatform_Detection(t *testing.T) {
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("CI_COMMIT_SHA", "sha")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	// Test that isMultiPlatform correctly detects multi-platform builds
	args := Args{
		Dockerfile: "Dockerfile",
		Platform:   "linux/amd64,linux/arm64",
	}
	assert.True(t, args.isMultiPlatform())
	assert.Equal(t, 2, args.platformCount())

	// Single platform should not be detected as multi-platform
	args.Platform = "linux/amd64"
	assert.False(t, args.isMultiPlatform())
	assert.Equal(t, 1, args.platformCount())
}

func TestArgs_IsMultiPlatform(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		want     bool
	}{
		{"empty", "", false},
		{"single platform", "linux/amd64", false},
		{"two platforms", "linux/amd64,linux/arm64", true},
		{"three platforms", "linux/amd64,linux/arm64,linux/arm/v7", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Args{Platform: tt.platform}
			assert.Equal(t, tt.want, a.isMultiPlatform())
		})
	}
}

func TestArgs_PlatformCount(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		want     int
	}{
		{"empty", "", 0},
		{"single platform", "linux/amd64", 1},
		{"two platforms", "linux/amd64,linux/arm64", 2},
		{"three platforms", "linux/amd64,linux/arm64,linux/arm/v7", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Args{Platform: tt.platform}
			assert.Equal(t, tt.want, a.platformCount())
		})
	}
}

func TestBuild_WithSkipLogin(t *testing.T) {
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("CI_COMMIT_SHA", "sha")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	client := &docker.MockDocker{}
	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    true,
		NoPull:     false,
	})
	assert.NoError(t, err)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Login <yellow>disabled</yellow>\n",
		"debug: Using build variables commit <green>sha</green> on branch <green>master</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:sha\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: sha\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
	})
}

func TestBuild_FeatureBranch(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	client := &docker.MockDocker{}
	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "build-tools-dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, 3, len(client.BuildOptions[0].BuildArgs))
	assert.Equal(t, "abc123", *client.BuildOptions[0].BuildArgs["CI_COMMIT"])
	assert.Equal(t, "feature1", *client.BuildOptions[0].BuildArgs["CI_BRANCH"])
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:feature1"}, client.BuildOptions[0].Tags)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:feature1"}, client.BuildOptions[0].Tags)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>feature1</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
	})
}

func TestBuild_MasterBranch(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "build-tools-dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[0].Tags)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>master</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
	})
}

func TestBuild_MainBranch(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "main")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "build-tools-dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:main", "repo/reponame:latest"}, client.BuildOptions[0].Tags)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>main</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:main\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: main\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:main\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
	})
}

func TestBuild_WithImageName(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "main")()
	defer pkg.SetEnv("IMAGE_NAME", "other")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")

	client := &docker.MockDocker{}
	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "build-tools-dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/other:abc123", "repo/other:main", "repo/other:latest"}, client.BuildOptions[0].Tags)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>main</green>\n",
		"info: Using other as BuildName\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/other:abc123\n    - repo/other:main\n    - repo/other:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: main\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/other:main\n    - repo/other:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
	})
}

func TestBuild_BadDockerHost(t *testing.T) {
	defer pkg.SetEnv("DOCKER_HOST", "abc-123")()
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	_, err := DoBuild(name, Args{})
	assert.EqualError(t, err, "unable to parse docker host `abc-123`")
	logMock.Check(t, []string{})
}

func TestBuild_Unreadable_Dockerfile(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	defer func() { _ = os.RemoveAll(name) }()
	dockerfile := filepath.Join(name, "Dockerfile")
	_ = os.MkdirAll(dockerfile, 0o777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}

	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.EqualError(t, err, fmt.Sprintf("read %s: is a directory", dockerfile))
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		fmt.Sprintf("error: <red>read %s: is a directory</red>", dockerfile),
	})
}

func TestBuild_Empty_Dockerfile(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "")

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}

	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.EqualError(t, err, "<red>the Dockerfile cannot be empty</red>")
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
	})
}

func TestBuild_Dockerfile_FromStdin(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "main")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}

	defer func() { _ = os.RemoveAll(name) }()
	err := os.MkdirAll(name, 0o777)
	assert.NoError(t, err)

	_, err = build(client, name, Args{
		Globals: args.Globals{
			StdIn: strings.NewReader("FROM scratch\n"),
		},
		Dockerfile: "-",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})
	assert.NoError(t, err)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"info: <greed>reading Dockerfile content from stdin</green>\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>main</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:main\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: main\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:main\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
	})
}

func TestBuild_HandleCaching(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}

	dockerfile := `
FROM scratch as build
RUN echo apa > file
FROM scratch as test
RUN echo cepa > file2
FROM scratch
COPY --from=build file .
COPY --from=test file2 .
`
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", dockerfile)

	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "build-tools-dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, 3, len(client.BuildOptions))
	assert.Equal(t, []string{"repo/reponame:build"}, client.BuildOptions[0].Tags)
	assert.Equal(t, []string{"repo/reponame:test"}, client.BuildOptions[1].Tags)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[2].Tags)
	assert.Equal(t, []string{"repo/reponame:build", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[0].CacheFrom)
	assert.Equal(t, []string{"repo/reponame:test", "repo/reponame:build", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[1].CacheFrom)
	assert.Equal(t, []string{"repo/reponame:test", "repo/reponame:build", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[2].CacheFrom)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>master</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:build\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: build\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:test\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: test\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
	})
}

func TestBuild_BrokenStage(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{BuildError: []error{nil, errors.New("build error")}}
	dockerfile := `
FROM scratch as build
RUN echo apa > file
FROM scratch as test
RUN echo cepa > file2
FROM scratch
COPY --from=build file .
COPY --from=test file2 .
`
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", dockerfile)
	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.EqualError(t, err, "build error")
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>master</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:build\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: build\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:test\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: test\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
	})
}

func TestBuild_ExportStage(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	dockerfile := `
FROM scratch as build
RUN echo apa > file
FROM scratch as test
RUN echo cepa > file2
FROM scratch as export
COPY --from=build file .
COPY --from=test file2 .
FROM scratch
COPY --from=build file .
COPY --from=test file2 .
`
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", dockerfile)
	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "build-tools-dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, 4, len(client.BuildOptions))
	assert.Equal(t, []string{"repo/reponame:build"}, client.BuildOptions[0].Tags)
	assert.Equal(t, []string{"repo/reponame:test"}, client.BuildOptions[1].Tags)
	assert.Equal(t, []string{"repo/reponame:export"}, client.BuildOptions[2].Tags)
	assert.Equal(t, []dockerbuild.ImageBuildOutput{
		{
			Type:  "local",
			Attrs: map[string]string{},
		},
	}, client.BuildOptions[2].Outputs)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[3].Tags)
	assert.Equal(t, []string{"repo/reponame:build", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[0].CacheFrom)
	assert.Equal(t, []string{"repo/reponame:test", "repo/reponame:build", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[1].CacheFrom)
	assert.Equal(t, []string{"repo/reponame:export", "repo/reponame:test", "repo/reponame:build", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[2].CacheFrom)
	assert.Equal(t, []string{"repo/reponame:export", "repo/reponame:test", "repo/reponame:build", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[3].CacheFrom)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>master</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:build\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: build\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:test\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: test\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:export\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:export\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: export\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs:\n    - type: local\n      attrs: {}\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:export\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
	})
}

func TestBuild_ExportAsLastStage(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	dockerfile := `
FROM scratch as build
RUN echo apa > file
FROM scratch as test
RUN echo cepa > file2
FROM scratch as export
COPY --from=build file .
COPY --from=test file2 .
`
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", dockerfile)
	_, err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "build-tools-dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, 4, len(client.BuildOptions))
	assert.Equal(t, []string{"repo/reponame:build"}, client.BuildOptions[0].Tags)
	assert.Equal(t, []string{"repo/reponame:test"}, client.BuildOptions[1].Tags)
	assert.Equal(t, []string{"repo/reponame:export"}, client.BuildOptions[2].Tags)
	assert.Equal(t, []dockerbuild.ImageBuildOutput{
		{
			Type:  "local",
			Attrs: map[string]string{},
		},
	}, client.BuildOptions[2].Outputs)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[3].Tags)
	assert.Equal(t, []string{"repo/reponame:build", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[0].CacheFrom)
	assert.Equal(t, []string{"repo/reponame:test", "repo/reponame:build", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[1].CacheFrom)
	assert.Equal(t, []string{"repo/reponame:export", "repo/reponame:test", "repo/reponame:build", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[2].CacheFrom)
	assert.Equal(t, []string{"repo/reponame:export", "repo/reponame:test", "repo/reponame:build", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[3].CacheFrom)
	logMock.Check(t, []string{
		"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>master</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:build\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: build\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:test\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: test\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:export\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:export\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: export\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs:\n    - type: local\n      attrs: {}\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: build-tools-dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:export\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
	})
}

type brokenReader struct{}

func (b brokenReader) Read([]byte) (n int, err error) {
	return 0, errors.New("read error")
}

var _ io.Reader = &brokenReader{}

func write(dir, file, content string) error {
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, file)), 0o777); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, file), []byte(fmt.Sprintln(strings.TrimSpace(content))), 0o666)
}

func TestBuild_MultiPlatform_ArgsValidation(t *testing.T) {
	// Test multi-platform argument validation and detection
	tests := []struct {
		name            string
		platform        string
		isMultiPlatform bool
		platformCount   int
	}{
		{"empty platform", "", false, 0},
		{"single platform linux/amd64", "linux/amd64", false, 1},
		{"single platform linux/arm64", "linux/arm64", false, 1},
		{"two platforms", "linux/amd64,linux/arm64", true, 2},
		{"three platforms", "linux/amd64,linux/arm64,linux/arm/v7", true, 3},
		{"platform with spaces", "linux/amd64, linux/arm64", true, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Args{Platform: tt.platform}
			assert.Equal(t, tt.isMultiPlatform, a.isMultiPlatform(), "isMultiPlatform mismatch")
			assert.Equal(t, tt.platformCount, a.platformCount(), "platformCount mismatch")
		})
	}
}

func Test_getBuildSharedKey(t *testing.T) {
	// Test that getBuildSharedKey returns consistent results for the same directory
	key1 := getBuildSharedKey("/test/dir")
	key2 := getBuildSharedKey("/test/dir")
	assert.Equal(t, key1, key2, "Same directory should produce same key")

	// Different directories should produce different keys
	key3 := getBuildSharedKey("/different/dir")
	assert.NotEqual(t, key1, key3, "Different directories should produce different keys")

	// Key should be a valid hex string
	assert.Regexp(t, "^[a-f0-9]+$", key1, "Key should be a hex string")
}

func Test_tryNodeIdentifier(t *testing.T) {
	// Test that tryNodeIdentifier returns a non-empty string
	id := tryNodeIdentifier()
	assert.NotEmpty(t, id, "Node identifier should not be empty")

	// Test that it returns consistent results
	id2 := tryNodeIdentifier()
	assert.Equal(t, id, id2, "Node identifier should be consistent")
}

func Test_provideSession(t *testing.T) {
	// Temporarily restore the original setupSession for this test
	originalSetupSession := setupSession
	setupSession = provideSession
	defer func() { setupSession = originalSetupSession }()

	// Test that provideSession returns a valid session
	session := provideSession("/test/dir")
	assert.NotNil(t, session, "Session should not be nil")
}

func TestArgs_DockerfileName(t *testing.T) {
	tests := []struct {
		name       string
		dockerfile string
		want       string
	}{
		{"normal dockerfile", "Dockerfile", "Dockerfile"},
		{"custom dockerfile", "Dockerfile.prod", "Dockerfile.prod"},
		{"stdin", "-", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Args{Dockerfile: tt.dockerfile}
			assert.Equal(t, tt.want, a.dockerfileName())
		})
	}
}

func TestArgs_IsDockerfileFromStdin(t *testing.T) {
	tests := []struct {
		name       string
		dockerfile string
		want       bool
	}{
		{"normal dockerfile", "Dockerfile", false},
		{"stdin", "-", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Args{Dockerfile: tt.dockerfile}
			assert.Equal(t, tt.want, a.isDockerfileFromStdin())
		})
	}
}

func Test_buildFrontendAttrs(t *testing.T) {
	tests := []struct {
		name       string
		dockerfile string
		platform   string
		target     string
		buildArgs  map[string]*string
		want       map[string]string
	}{
		{
			name:       "basic attributes",
			dockerfile: "Dockerfile",
			platform:   "linux/amd64",
			target:     "",
			buildArgs:  nil,
			want: map[string]string{
				"filename": "Dockerfile",
				"platform": "linux/amd64",
			},
		},
		{
			name:       "with build args",
			dockerfile: "Dockerfile.prod",
			platform:   "linux/amd64,linux/arm64",
			target:     "",
			buildArgs: map[string]*string{
				"VERSION": strPtr("1.0.0"),
				"DEBUG":   strPtr("true"),
			},
			want: map[string]string{
				"filename":          "Dockerfile.prod",
				"platform":          "linux/amd64,linux/arm64",
				"build-arg:VERSION": "1.0.0",
				"build-arg:DEBUG":   "true",
			},
		},
		{
			name:       "with nil build arg value",
			dockerfile: "Dockerfile",
			platform:   "linux/arm64",
			target:     "",
			buildArgs: map[string]*string{
				"SET":   strPtr("value"),
				"UNSET": nil,
			},
			want: map[string]string{
				"filename":      "Dockerfile",
				"platform":      "linux/arm64",
				"build-arg:SET": "value",
			},
		},
		{
			name:       "empty platform",
			dockerfile: "Dockerfile",
			platform:   "",
			target:     "",
			buildArgs:  nil,
			want: map[string]string{
				"filename": "Dockerfile",
			},
		},
		{
			name:       "with target stage",
			dockerfile: "Dockerfile",
			platform:   "linux/amd64",
			target:     "builder",
			buildArgs:  nil,
			want: map[string]string{
				"filename": "Dockerfile",
				"platform": "linux/amd64",
				"target":   "builder",
			},
		},
		{
			name:       "with target and build args",
			dockerfile: "Dockerfile",
			platform:   "linux/amd64,linux/arm64",
			target:     "production",
			buildArgs: map[string]*string{
				"VERSION": strPtr("2.0.0"),
			},
			want: map[string]string{
				"filename":          "Dockerfile",
				"platform":          "linux/amd64,linux/arm64",
				"target":            "production",
				"build-arg:VERSION": "2.0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildFrontendAttrs(tt.dockerfile, tt.platform, tt.target, tt.buildArgs)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_buildExportEntry(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want string
	}{
		{
			name: "single tag",
			tags: []string{"registry.example.com/image:v1"},
			want: "registry.example.com/image:v1",
		},
		{
			name: "multiple tags",
			tags: []string{
				"registry.example.com/image:v1",
				"registry.example.com/image:latest",
				"registry.example.com/image:sha-abc123",
			},
			want: "registry.example.com/image:v1,registry.example.com/image:latest,registry.example.com/image:sha-abc123",
		},
		{
			name: "empty tags",
			tags: []string{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := buildExportEntry(tt.tags)
			assert.Equal(t, "image", entry.Type)
			assert.Equal(t, tt.want, entry.Attrs["name"])
			assert.Equal(t, "true", entry.Attrs["push"])
		})
	}
}

func Test_buildCacheImports(t *testing.T) {
	tests := []struct {
		name     string
		caches   []string
		ecrCache *config.ECRCache
		want     int
	}{
		{
			name:     "no caches",
			caches:   []string{},
			ecrCache: nil,
			want:     0,
		},
		{
			name:     "nil caches",
			caches:   nil,
			ecrCache: nil,
			want:     0,
		},
		{
			name:     "single cache",
			caches:   []string{"registry.example.com/image:cache"},
			ecrCache: nil,
			want:     1,
		},
		{
			name: "multiple caches",
			caches: []string{
				"registry.example.com/image:branch",
				"registry.example.com/image:latest",
			},
			ecrCache: nil,
			want:     2,
		},
		{
			name:     "with ECR cache",
			caches:   []string{"registry.example.com/image:cache"},
			ecrCache: &config.ECRCache{Url: "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache"},
			want:     2,
		},
		{
			name:     "only ECR cache",
			caches:   nil,
			ecrCache: &config.ECRCache{Url: "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache", Tag: "custom-tag"},
			want:     1,
		},
		{
			name:     "empty ECR cache",
			caches:   []string{"registry.example.com/image:cache"},
			ecrCache: &config.ECRCache{},
			want:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := buildCacheImports(tt.caches, tt.ecrCache)
			assert.Len(t, imports, tt.want)

			// ECR cache should be first if configured
			offset := 0
			if tt.ecrCache != nil && tt.ecrCache.Configured() {
				firstImport := imports[0]
				assert.Equal(t, "registry", firstImport.Type)
				assert.Equal(t, tt.ecrCache.CacheRef(), firstImport.Attrs["ref"])
				offset = 1
			}

			// Verify regular caches come after ECR cache
			for i, cache := range tt.caches {
				assert.Equal(t, "registry", imports[offset+i].Type)
				assert.Equal(t, cache, imports[offset+i].Attrs["ref"])
			}
		})
	}
}

func Test_buildCacheExports(t *testing.T) {
	tests := []struct {
		name     string
		ecrCache *config.ECRCache
		want     int
	}{
		{
			name:     "nil ECR cache",
			ecrCache: nil,
			want:     0,
		},
		{
			name:     "empty ECR cache",
			ecrCache: &config.ECRCache{},
			want:     0,
		},
		{
			name:     "configured ECR cache with default tag",
			ecrCache: &config.ECRCache{Url: "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache"},
			want:     1,
		},
		{
			name:     "configured ECR cache with custom tag",
			ecrCache: &config.ECRCache{Url: "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache", Tag: "custom-tag"},
			want:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exports := buildCacheExports(tt.ecrCache)
			assert.Len(t, exports, tt.want)

			if tt.want > 0 {
				export := exports[0]
				assert.Equal(t, "registry", export.Type)
				assert.Equal(t, tt.ecrCache.CacheRef(), export.Attrs["ref"])
				assert.Equal(t, "max", export.Attrs["mode"])
				assert.Equal(t, "true", export.Attrs["image-manifest"])
				assert.Equal(t, "true", export.Attrs["oci-mediatypes"])
			}
		})
	}
}

func Test_hasContainerdSnapshotter(t *testing.T) {
	tests := []struct {
		name    string
		workers []*client.WorkerInfo
		want    bool
	}{
		{
			name:    "nil workers",
			workers: nil,
			want:    false,
		},
		{
			name:    "empty workers",
			workers: []*client.WorkerInfo{},
			want:    false,
		},
		{
			name: "worker without containerd label",
			workers: []*client.WorkerInfo{
				{Labels: map[string]string{"foo": "bar"}},
			},
			want: false,
		},
		{
			name: "worker with containerd in label key",
			workers: []*client.WorkerInfo{
				{Labels: map[string]string{"containerd.snapshotter": "overlayfs"}},
			},
			want: true,
		},
		{
			name: "worker with containerd in label value",
			workers: []*client.WorkerInfo{
				{Labels: map[string]string{"snapshotter": "containerd"}},
			},
			want: true,
		},
		{
			name: "multiple workers one with containerd",
			workers: []*client.WorkerInfo{
				{Labels: map[string]string{"foo": "bar"}},
				{Labels: map[string]string{"snapshotter": "containerd-overlayfs"}},
			},
			want: true,
		},
		{
			name: "worker with empty labels",
			workers: []*client.WorkerInfo{
				{Labels: map[string]string{}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasContainerdSnapshotter(tt.workers)
			assert.Equal(t, tt.want, got)
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func Test_buildMultiPlatformWithFactory_Success(t *testing.T) {
	defer pkg.SetEnv("BUILDKIT_HOST", "tcp://localhost:1234")()

	var capturedOpts client.SolveOpt
	mockClient := &MockBuildkitClient{
		SolveFunc: func(ctx context.Context, def *llb.Definition, opt client.SolveOpt, statusChan chan *client.SolveStatus) (*client.SolveResponse, error) {
			capturedOpts = opt
			close(statusChan)
			return &client.SolveResponse{
				ExporterResponse: map[string]string{
					"containerimage.digest": "sha256:abc123def456",
				},
			}, nil
		},
	}

	mockFactory := func(ctx context.Context, address string, opts ...client.ClientOpt) (BuildkitClient, error) {
		assert.Equal(t, "tcp://localhost:1234", address)
		return mockClient, nil
	}

	dir := t.TempDir()
	_ = write(dir, "Dockerfile", "FROM scratch")

	digest, err := buildMultiPlatformWithFactory(
		&docker.MockDocker{},
		dir,
		Args{Platform: "linux/amd64,linux/arm64", Dockerfile: "Dockerfile"},
		map[string]*string{"VERSION": strPtr("1.0.0")},
		[]string{"registry.example.com/image:v1", "registry.example.com/image:latest"},
		[]string{"registry.example.com/image:cache"},
		"",
		nil,
		nil,
		mockFactory,
	)

	assert.NoError(t, err)
	assert.Equal(t, "sha256:abc123def456", digest)
	assert.Equal(t, "dockerfile.v0", capturedOpts.Frontend)
	assert.Equal(t, "linux/amd64,linux/arm64", capturedOpts.FrontendAttrs["platform"])
	assert.Equal(t, "1.0.0", capturedOpts.FrontendAttrs["build-arg:VERSION"])
	assert.Len(t, capturedOpts.Exports, 1)
	assert.Equal(t, "registry.example.com/image:v1,registry.example.com/image:latest", capturedOpts.Exports[0].Attrs["name"])
	assert.Len(t, capturedOpts.CacheImports, 1)
}

func Test_buildMultiPlatformWithFactory_ClientConnectionError(t *testing.T) {
	defer pkg.SetEnv("BUILDKIT_HOST", "tcp://localhost:1234")()

	mockFactory := func(ctx context.Context, address string, opts ...client.ClientOpt) (BuildkitClient, error) {
		return nil, errors.New("connection refused")
	}

	dir := t.TempDir()
	_ = write(dir, "Dockerfile", "FROM scratch")

	_, err := buildMultiPlatformWithFactory(
		&docker.MockDocker{},
		dir,
		Args{Platform: "linux/amd64,linux/arm64", Dockerfile: "Dockerfile"},
		nil,
		[]string{"registry.example.com/image:v1"},
		nil,
		"",
		nil,
		nil,
		mockFactory,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to buildkit")
	assert.Contains(t, err.Error(), "connection refused")
}

func Test_buildMultiPlatformWithFactory_SolveError(t *testing.T) {
	defer pkg.SetEnv("BUILDKIT_HOST", "tcp://localhost:1234")()

	mockClient := &MockBuildkitClient{
		SolveFunc: func(ctx context.Context, def *llb.Definition, opt client.SolveOpt, statusChan chan *client.SolveStatus) (*client.SolveResponse, error) {
			close(statusChan)
			return nil, errors.New("build failed")
		},
	}

	mockFactory := func(ctx context.Context, address string, opts ...client.ClientOpt) (BuildkitClient, error) {
		return mockClient, nil
	}

	dir := t.TempDir()
	_ = write(dir, "Dockerfile", "FROM scratch")

	_, err := buildMultiPlatformWithFactory(
		&docker.MockDocker{},
		dir,
		Args{Platform: "linux/amd64,linux/arm64", Dockerfile: "Dockerfile"},
		nil,
		[]string{"registry.example.com/image:v1"},
		nil,
		"",
		nil,
		nil,
		mockFactory,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "multi-platform build failed")
}

func Test_buildMultiPlatformWithFactory_ExporterNotFoundError(t *testing.T) {
	defer pkg.SetEnv("BUILDKIT_HOST", "tcp://localhost:1234")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)

	mockClient := &MockBuildkitClient{
		SolveFunc: func(ctx context.Context, def *llb.Definition, opt client.SolveOpt, statusChan chan *client.SolveStatus) (*client.SolveResponse, error) {
			close(statusChan)
			return nil, errors.New("exporter 'image' could not be found")
		},
	}

	mockFactory := func(ctx context.Context, address string, opts ...client.ClientOpt) (BuildkitClient, error) {
		return mockClient, nil
	}

	dir := t.TempDir()
	_ = write(dir, "Dockerfile", "FROM scratch")

	_, err := buildMultiPlatformWithFactory(
		&docker.MockDocker{},
		dir,
		Args{Platform: "linux/amd64,linux/arm64", Dockerfile: "Dockerfile"},
		nil,
		[]string{"registry.example.com/image:v1"},
		nil,
		"",
		nil,
		nil,
		mockFactory,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "multi-platform build failed")
	// Check that helpful error messages were logged
	found := false
	for _, entry := range logMock.Logged {
		if strings.Contains(entry, "containerd-snapshotter") {
			found = true
			break
		}
	}
	assert.True(t, found, "Should log containerd snapshotter help message")
}

func Test_buildMultiPlatformWithFactory_InvalidDirectory(t *testing.T) {
	defer pkg.SetEnv("BUILDKIT_HOST", "tcp://localhost:1234")()

	mockFactory := func(ctx context.Context, address string, opts ...client.ClientOpt) (BuildkitClient, error) {
		return &MockBuildkitClient{}, nil
	}

	_, err := buildMultiPlatformWithFactory(
		&docker.MockDocker{},
		"/nonexistent/directory/that/does/not/exist",
		Args{Platform: "linux/amd64,linux/arm64", Dockerfile: "Dockerfile"},
		nil,
		[]string{"registry.example.com/image:v1"},
		nil,
		"",
		nil,
		nil,
		mockFactory,
	)

	assert.Error(t, err)
}

func Test_buildMultiPlatformWithFactory_ViaDocker(t *testing.T) {
	// Test the Docker daemon path (no BUILDKIT_HOST)
	defer pkg.SetEnv("BUILDKIT_HOST", "")()

	var capturedOpts []client.ClientOpt
	mockClient := &MockBuildkitClient{
		ListWorkersFunc: func(ctx context.Context, opts ...client.ListWorkersOption) ([]*client.WorkerInfo, error) {
			return []*client.WorkerInfo{
				{Labels: map[string]string{"containerd.snapshotter": "overlayfs"}},
			}, nil
		},
		SolveFunc: func(ctx context.Context, def *llb.Definition, opt client.SolveOpt, statusChan chan *client.SolveStatus) (*client.SolveResponse, error) {
			close(statusChan)
			return &client.SolveResponse{}, nil
		},
	}

	mockFactory := func(ctx context.Context, address string, opts ...client.ClientOpt) (BuildkitClient, error) {
		capturedOpts = opts
		assert.Equal(t, "", address) // Empty address when using Docker daemon
		return mockClient, nil
	}

	dir := t.TempDir()
	_ = write(dir, "Dockerfile", "FROM scratch")

	_, err := buildMultiPlatformWithFactory(
		&docker.MockDocker{},
		dir,
		Args{Platform: "linux/amd64,linux/arm64", Dockerfile: "Dockerfile"},
		nil,
		[]string{"registry.example.com/image:v1"},
		nil,
		"",
		nil,
		nil,
		mockFactory,
	)

	assert.NoError(t, err)
	// Should have context and session dialers when using Docker daemon
	assert.Len(t, capturedOpts, 2)
}

func Test_buildMultiPlatformWithFactory_ListWorkersError(t *testing.T) {
	defer pkg.SetEnv("BUILDKIT_HOST", "")()

	mockClient := &MockBuildkitClient{
		ListWorkersFunc: func(ctx context.Context, opts ...client.ListWorkersOption) ([]*client.WorkerInfo, error) {
			return nil, errors.New("failed to list workers")
		},
	}

	mockFactory := func(ctx context.Context, address string, opts ...client.ClientOpt) (BuildkitClient, error) {
		return mockClient, nil
	}

	dir := t.TempDir()
	_ = write(dir, "Dockerfile", "FROM scratch")

	_, err := buildMultiPlatformWithFactory(
		&docker.MockDocker{},
		dir,
		Args{Platform: "linux/amd64,linux/arm64", Dockerfile: "Dockerfile"},
		nil,
		[]string{"registry.example.com/image:v1"},
		nil,
		"",
		nil,
		nil,
		mockFactory,
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list buildkit workers")
}

func Test_buildMultiPlatformWithFactory_WithECRCache(t *testing.T) {
	defer pkg.SetEnv("BUILDKIT_HOST", "tcp://localhost:1234")()

	var capturedOpts client.SolveOpt
	mockClient := &MockBuildkitClient{
		SolveFunc: func(ctx context.Context, def *llb.Definition, opt client.SolveOpt, statusChan chan *client.SolveStatus) (*client.SolveResponse, error) {
			capturedOpts = opt
			close(statusChan)
			return &client.SolveResponse{}, nil
		},
	}

	mockFactory := func(ctx context.Context, address string, opts ...client.ClientOpt) (BuildkitClient, error) {
		return mockClient, nil
	}

	dir := t.TempDir()
	_ = write(dir, "Dockerfile", "FROM scratch")

	ecrCache := &config.ECRCache{
		Url: "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache",
		Tag: "my-cache",
	}

	_, err := buildMultiPlatformWithFactory(
		&docker.MockDocker{},
		dir,
		Args{Platform: "linux/amd64,linux/arm64", Dockerfile: "Dockerfile"},
		nil,
		[]string{"registry.example.com/image:v1"},
		[]string{"registry.example.com/image:branch"},
		"",
		ecrCache,
		nil,
		mockFactory,
	)

	assert.NoError(t, err)

	// Verify cache imports include ECR cache (ECR cache is first for priority)
	assert.Len(t, capturedOpts.CacheImports, 2)
	assert.Equal(t, "registry", capturedOpts.CacheImports[0].Type)
	assert.Equal(t, "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache:my-cache", capturedOpts.CacheImports[0].Attrs["ref"])
	assert.Equal(t, "registry", capturedOpts.CacheImports[1].Type)
	assert.Equal(t, "registry.example.com/image:branch", capturedOpts.CacheImports[1].Attrs["ref"])

	// Verify cache exports have ECR-specific settings
	assert.Len(t, capturedOpts.CacheExports, 1)
	assert.Equal(t, "registry", capturedOpts.CacheExports[0].Type)
	assert.Equal(t, "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache:my-cache", capturedOpts.CacheExports[0].Attrs["ref"])
	assert.Equal(t, "max", capturedOpts.CacheExports[0].Attrs["mode"])
	assert.Equal(t, "true", capturedOpts.CacheExports[0].Attrs["image-manifest"])
	assert.Equal(t, "true", capturedOpts.CacheExports[0].Attrs["oci-mediatypes"])
}

func Test_buildMultiPlatformWithFactory_WithECRCache_DefaultTag(t *testing.T) {
	defer pkg.SetEnv("BUILDKIT_HOST", "tcp://localhost:1234")()

	var capturedOpts client.SolveOpt
	mockClient := &MockBuildkitClient{
		SolveFunc: func(ctx context.Context, def *llb.Definition, opt client.SolveOpt, statusChan chan *client.SolveStatus) (*client.SolveResponse, error) {
			capturedOpts = opt
			close(statusChan)
			return &client.SolveResponse{}, nil
		},
	}

	mockFactory := func(ctx context.Context, address string, opts ...client.ClientOpt) (BuildkitClient, error) {
		return mockClient, nil
	}

	dir := t.TempDir()
	_ = write(dir, "Dockerfile", "FROM scratch")

	ecrCache := &config.ECRCache{
		Url: "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache",
	}

	_, err := buildMultiPlatformWithFactory(
		&docker.MockDocker{},
		dir,
		Args{Platform: "linux/amd64,linux/arm64", Dockerfile: "Dockerfile"},
		nil,
		[]string{"registry.example.com/image:v1"},
		nil,
		"",
		ecrCache,
		nil,
		mockFactory,
	)

	assert.NoError(t, err)

	// Verify cache imports include ECR cache with default tag
	assert.Len(t, capturedOpts.CacheImports, 1)
	assert.Equal(t, "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache:buildcache", capturedOpts.CacheImports[0].Attrs["ref"])

	// Verify cache exports
	assert.Len(t, capturedOpts.CacheExports, 1)
	assert.Equal(t, "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache:buildcache", capturedOpts.CacheExports[0].Attrs["ref"])
}

func Test_buildMultiPlatformWithFactory_ExportStage(t *testing.T) {
	defer pkg.SetEnv("BUILDKIT_HOST", "tcp://localhost:1234")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)

	var capturedOpts client.SolveOpt
	mockClient := &MockBuildkitClient{
		SolveFunc: func(ctx context.Context, def *llb.Definition, opt client.SolveOpt, statusChan chan *client.SolveStatus) (*client.SolveResponse, error) {
			capturedOpts = opt
			close(statusChan)
			return &client.SolveResponse{}, nil
		},
	}

	mockFactory := func(ctx context.Context, address string, opts ...client.ClientOpt) (BuildkitClient, error) {
		return mockClient, nil
	}

	dir := t.TempDir()
	_ = write(dir, "Dockerfile", "FROM scratch AS export-artifacts\nCOPY . /out")

	_, err := buildMultiPlatformWithFactory(
		&docker.MockDocker{},
		dir,
		Args{Platform: "linux/amd64", Dockerfile: "Dockerfile"},
		nil,
		[]string{"registry.example.com/image:v1"},
		nil,
		"export-artifacts", // export stage target
		nil,
		nil,
		mockFactory,
	)

	assert.NoError(t, err)

	// Verify that the export uses local exporter instead of image exporter
	assert.Len(t, capturedOpts.Exports, 1)
	assert.Equal(t, client.ExporterLocal, capturedOpts.Exports[0].Type)
	assert.Equal(t, filepath.Join(dir, "exported"), capturedOpts.Exports[0].OutputDir)

	// Verify target is set in frontend attrs
	assert.Equal(t, "export-artifacts", capturedOpts.FrontendAttrs["target"])

	// Verify the export directory was created
	_, err = os.Stat(filepath.Join(dir, "exported"))
	assert.NoError(t, err)

	// Verify log message about exporting
	foundExportLog := false
	for _, entry := range logMock.Logged {
		if strings.Contains(entry, "Exporting build artifacts") {
			foundExportLog = true
			break
		}
	}
	assert.True(t, foundExportLog, "Should log message about exporting build artifacts")
}

func Test_buildMultiPlatformWithFactory_ExportStage_NoAuthWarning(t *testing.T) {
	defer pkg.SetEnv("BUILDKIT_HOST", "tcp://localhost:1234")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)

	mockClient := &MockBuildkitClient{
		SolveFunc: func(ctx context.Context, def *llb.Definition, opt client.SolveOpt, statusChan chan *client.SolveStatus) (*client.SolveResponse, error) {
			close(statusChan)
			return &client.SolveResponse{}, nil
		},
	}

	mockFactory := func(ctx context.Context, address string, opts ...client.ClientOpt) (BuildkitClient, error) {
		return mockClient, nil
	}

	dir := t.TempDir()
	_ = write(dir, "Dockerfile", "FROM scratch AS export-files\nCOPY . /out")

	// Call without authenticator - should NOT warn for export stages
	_, err := buildMultiPlatformWithFactory(
		&docker.MockDocker{},
		dir,
		Args{Platform: "linux/amd64", Dockerfile: "Dockerfile"},
		nil,
		[]string{"registry.example.com/image:v1"},
		nil,
		"export-files", // export stage target
		nil,
		nil, // no authenticator
		mockFactory,
	)

	assert.NoError(t, err)

	// Verify NO warning about missing authenticator for export stages
	for _, entry := range logMock.Logged {
		assert.NotContains(t, entry, "No authenticator provided", "Should not warn about missing authenticator for export stages")
	}
}

func Test_buildMultiPlatformWithFactory_NonExportStage_WarnNoAuth(t *testing.T) {
	defer pkg.SetEnv("BUILDKIT_HOST", "tcp://localhost:1234")()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)

	mockClient := &MockBuildkitClient{
		SolveFunc: func(ctx context.Context, def *llb.Definition, opt client.SolveOpt, statusChan chan *client.SolveStatus) (*client.SolveResponse, error) {
			close(statusChan)
			return &client.SolveResponse{}, nil
		},
	}

	mockFactory := func(ctx context.Context, address string, opts ...client.ClientOpt) (BuildkitClient, error) {
		return mockClient, nil
	}

	dir := t.TempDir()
	_ = write(dir, "Dockerfile", "FROM scratch")

	// Call without authenticator for non-export stage - should warn
	_, err := buildMultiPlatformWithFactory(
		&docker.MockDocker{},
		dir,
		Args{Platform: "linux/amd64", Dockerfile: "Dockerfile"},
		nil,
		[]string{"registry.example.com/image:v1"},
		nil,
		"", // empty target - regular build
		nil,
		nil, // no authenticator
		mockFactory,
	)

	assert.NoError(t, err)

	// Verify warning about missing authenticator for non-export stages
	foundWarning := false
	for _, entry := range logMock.Logged {
		if strings.Contains(entry, "No authenticator provided") {
			foundWarning = true
			break
		}
	}
	assert.True(t, foundWarning, "Should warn about missing authenticator for non-export stages")
}

func Test_extractHost(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "ECR URL with repo",
			url:  "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache-repo",
			want: "123456789012.dkr.ecr.us-east-1.amazonaws.com",
		},
		{
			name: "ECR URL without repo",
			url:  "123456789012.dkr.ecr.us-east-1.amazonaws.com",
			want: "123456789012.dkr.ecr.us-east-1.amazonaws.com",
		},
		{
			name: "URL with https prefix",
			url:  "https://123456789012.dkr.ecr.us-east-1.amazonaws.com/cache-repo",
			want: "123456789012.dkr.ecr.us-east-1.amazonaws.com",
		},
		{
			name: "URL with http prefix",
			url:  "http://registry.example.com/repo",
			want: "registry.example.com",
		},
		{
			name: "Docker Hub style",
			url:  "docker.io/library/alpine",
			want: "docker.io",
		},
		{
			name: "GitLab registry",
			url:  "registry.gitlab.com/org/project",
			want: "registry.gitlab.com",
		},
		{
			name: "empty string",
			url:  "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHost(tt.url)
			assert.Equal(t, tt.want, got)
		})
	}
}
