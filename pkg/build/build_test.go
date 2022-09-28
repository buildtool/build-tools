// MIT License
//
// Copyright (c) 2021 buildtool
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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apex/log"
	"github.com/docker/docker/api/types"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/stretchr/testify/assert"

	"github.com/buildtool/build-tools/pkg"
	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/docker"
)

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
	_ = os.WriteFile(file, []byte(yaml), 0777)
	defer func() { _ = os.Remove(file) }()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}
	defer func() { _ = os.RemoveAll(name) }()
	_ = write(name, "Dockerfile", "FROM scratch")
	err := build(client, name, Args{
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

	err := DoBuild(name, Args{Dockerfile: "Dockerfile"})
	assert.NoError(t, err)
	logMock.Check(t, []string{"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>No docker registry</green>\n",
		"debug: Authenticating against registry <green>No docker registry</green>\n",
		"debug: Authentication <yellow>not supported</yellow> for registry <green>No docker registry</green>\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>feature1</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - noregistry/reponame:abc123\n    - noregistry/reponame:feature1\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - noregistry/reponame:feature1\n    - noregistry/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
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
	err := build(client, name, Args{
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
		"error: Unable to login\n"})
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
	err := build(client, name, Args{
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
	err := build(client, name, Args{
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
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
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
	err := build(client, name, Args{
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
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
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
	err := build(client, name, Args{
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
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
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
	err = build(client, name, Args{
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
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
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
	err := build(client, name, Args{
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
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"error: failed to parse aux message: json: cannot unmarshal number into Go struct field BuildResult.ID of type string",
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

	err := build(client, name, Args{
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

	err := build(client, name, Args{
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
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:sha\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: sha\n    buildargs1: 1=1\n    buildargs4: env-value\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
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

	err := build(client, name, Args{
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
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:sha\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: sha\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: linux/amd64\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
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
	err := build(client, name, Args{
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
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:sha\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: sha\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
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
	err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, 3, len(client.BuildOptions[0].BuildArgs))
	assert.Equal(t, "abc123", *client.BuildOptions[0].BuildArgs["CI_COMMIT"])
	assert.Equal(t, "feature1", *client.BuildOptions[0].BuildArgs["CI_BRANCH"])
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:feature1"}, client.BuildOptions[0].Tags)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:feature1"}, client.BuildOptions[0].Tags)
	logMock.Check(t, []string{"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>feature1</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:feature1\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: feature1\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:feature1\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
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

	err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[0].Tags)
	logMock.Check(t, []string{"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>master</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
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

	err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:main", "repo/reponame:latest"}, client.BuildOptions[0].Tags)
	logMock.Check(t, []string{"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>main</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:main\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: main\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:main\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
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
	err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/other:abc123", "repo/other:main", "repo/other:latest"}, client.BuildOptions[0].Tags)
	logMock.Check(t, []string{"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>main</green>\n",
		"info: Using other as BuildName\n",
		"info: Using other as BuildName\n",
		"info: Using other as BuildName\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/other:abc123\n    - repo/other:main\n    - repo/other:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: main\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/other:main\n    - repo/other:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
}

func TestBuild_BadDockerHost(t *testing.T) {
	defer pkg.SetEnv("DOCKER_HOST", "abc-123")()
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := DoBuild(name, Args{})
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
	_ = os.MkdirAll(dockerfile, 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	client := &docker.MockDocker{}

	err := build(client, name, Args{
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

	err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
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
	logMock.Check(t, []string{"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>master</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:build\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: build\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:test\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: test\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
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
	err := build(client, name, Args{
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
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:build\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: build\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:test\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: test\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
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
	err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, 4, len(client.BuildOptions))
	assert.Equal(t, []string{"repo/reponame:build"}, client.BuildOptions[0].Tags)
	assert.Equal(t, []string{"repo/reponame:test"}, client.BuildOptions[1].Tags)
	assert.Equal(t, []string{"repo/reponame:export"}, client.BuildOptions[2].Tags)
	assert.Equal(t, []types.ImageBuildOutput{
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
	logMock.Check(t, []string{"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>master</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:build\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: build\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:test\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: test\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:export\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:export\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: export\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs:\n    - type: local\n      attrs: {}\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:export\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
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
	err := build(client, name, Args{
		Globals:    args.Globals{},
		Dockerfile: "Dockerfile",
		BuildArgs:  nil,
		NoLogin:    false,
		NoPull:     false,
	})

	assert.NoError(t, err)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, 4, len(client.BuildOptions))
	assert.Equal(t, []string{"repo/reponame:build"}, client.BuildOptions[0].Tags)
	assert.Equal(t, []string{"repo/reponame:test"}, client.BuildOptions[1].Tags)
	assert.Equal(t, []string{"repo/reponame:export"}, client.BuildOptions[2].Tags)
	assert.Equal(t, []types.ImageBuildOutput{
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
	logMock.Check(t, []string{"debug: Using CI <green>Gitlab</green>\n",
		"debug: Using registry <green>Dockerhub</green>\n",
		"debug: Authenticating against registry <green>Dockerhub</green>\n",
		"debug: Logged in\n",
		"debug: Using build variables commit <green>abc123</green> on branch <green>master</green>\n",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:build\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: build\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:test\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: test\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:export\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:export\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: export\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs:\n    - type: local\n      attrs: {}\n\n",
		"info: Build successful",
		"debug: performing docker build with options (auths removed):\ntags:\n    - repo/reponame:abc123\n    - repo/reponame:master\n    - repo/reponame:latest\nsuppressoutput: false\nremotecontext: client-session\nnocache: false\nremove: true\nforceremove: false\npullparent: true\nisolation: \"\"\ncpusetcpus: \"\"\ncpusetmems: \"\"\ncpushares: 0\ncpuquota: 0\ncpuperiod: 0\nmemory: 0\nmemoryswap: -1\ncgroupparent: \"\"\nnetworkmode: \"\"\nshmsize: 268435456\ndockerfile: Dockerfile\nulimits: []\nbuildargs:\n    BUILDKIT_INLINE_CACHE: \"1\"\n    CI_BRANCH: master\n    CI_COMMIT: abc123\nauthconfigs: {}\ncontext: null\nlabels: {}\nsquash: false\ncachefrom:\n    - repo/reponame:export\n    - repo/reponame:test\n    - repo/reponame:build\n    - repo/reponame:master\n    - repo/reponame:latest\nsecurityopt: []\nextrahosts: []\ntarget: \"\"\nsessionid: \"\"\nplatform: \"\"\nversion: \"2\"\nbuildid: \"\"\noutputs: []\n\n",
		"info: Build successful"})
}

type brokenReader struct{}

func (b brokenReader) Read([]byte) (n int, err error) {
	return 0, errors.New("read error")
}

var _ io.Reader = &brokenReader{}

func write(dir, file, content string) error {
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, file)), 0777); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, file), []byte(fmt.Sprintln(strings.TrimSpace(content))), 0666)
}
