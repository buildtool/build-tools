package build

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/buildtool/build-tools/pkg"
	"github.com/buildtool/build-tools/pkg/docker"
	"github.com/docker/docker/pkg/archive"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var name string

func TestMain(m *testing.M) {
	tempDir := setup()
	code := m.Run()
	teardown(tempDir)
	os.Exit(code)
}

func setup() string {
	name, _ = ioutil.TempDir(os.TempDir(), "build-tools")
	os.Clearenv()

	return name
}

func teardown(tempDir string) {
	_ = os.RemoveAll(tempDir)
}

func TestBuild_BrokenConfig(t *testing.T) {
	yaml := `ci: [] `
	file := filepath.Join(name, ".buildtools.yaml")
	_ = ioutil.WriteFile(file, []byte(yaml), 0777)
	defer func() { _ = os.Remove(file) }()

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), out, eout, "Dockerfile", arrayFlags{}, false, false)

	absPath, _ := filepath.Abs(filepath.Join(name, ".buildtools.yaml"))
	assert.Equal(t, -3, code)
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s'\x1b[39m\x1b[0m\n", absPath), out.String())
	assert.Equal(t, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig\n", eout.String())
}

func TestBuild_NoRegistry(t *testing.T) {
	filename := filepath.Join(name, "Dockerfile")
	_ = ioutil.WriteFile(filename, []byte("FROM scratch"), 0777)
	defer func() { _ = os.RemoveAll(filename) }()

	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	tmpDockerClient := dockerClient
	dockerClient = func() (client docker.Client, e error) {
		return &docker.MockDocker{}, nil
	}
	defer func() { dockerClient = tmpDockerClient }()

	code := DoBuild(name, out, eout)
	assert.Equal(t, 0, code)
	assert.Equal(t, "\x1b[0mUsing CI \x1b[32mGitlab\x1b[39m\x1b[0m\n\x1b[0mUsing registry \x1b[32mNo docker registry\x1b[39m\x1b[0m\n\x1b[0mAuthenticating against registry \x1b[32mNo docker registry\x1b[39m\x1b[0m\n\x1b[0mAuthentication \x1b[33mnot supported\x1b[39m for registry \x1b[32mNo docker registry\x1b[39m\x1b[0m\n\x1b[0mUsing build variables commit \x1b[32mabc123\x1b[39m on branch \x1b[32mfeature1\x1b[39m\x1b[0m\nBuild successful", out.String())
	assert.Equal(t, "", eout.String())
}

func TestBuild_LoginError(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), out, eout, "Dockerfile", arrayFlags{}, false, false)

	assert.Equal(t, -4, code)
	assert.Equal(t, "\x1b[0mUsing CI \x1b[32mGitlab\x1b[39m\x1b[0m\n\x1b[0mUsing registry \x1b[32mDockerhub\x1b[39m\x1b[0m\n\x1b[0mAuthenticating against registry \x1b[32mDockerhub\x1b[39m\x1b[0m\nUnable to login\n", out.String())
	assert.Equal(t, "invalid username/password\n", eout.String())
}

func TestBuild_BuildError(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{BuildError: []error{fmt.Errorf("build error")}}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), out, eout, "Dockerfile", arrayFlags{}, false, false)

	assert.Equal(t, -7, code)
	assert.Equal(t, "\x1b[0mUsing CI \x1b[32mGitlab\x1b[39m\x1b[0m\n\x1b[0mUsing registry \x1b[32mDockerhub\x1b[39m\x1b[0m\n\x1b[0mAuthenticating against registry \x1b[32mDockerhub\x1b[39m\x1b[0m\nLogged in\n\x1b[0mUsing build variables commit \x1b[32mabc123\x1b[39m on branch \x1b[32mfeature1\x1b[39m\x1b[0m\n", out.String())
	assert.Equal(t, "build error\n", eout.String())
}

func TestBuild_BuildResponseError(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{ResponseError: fmt.Errorf("build error")}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), out, eout, "Dockerfile", arrayFlags{}, false, false)

	assert.Equal(t, -7, code)
	assert.Equal(t, "\x1b[0mUsing CI \x1b[32mGitlab\x1b[39m\x1b[0m\n\x1b[0mUsing registry \x1b[32mDockerhub\x1b[39m\x1b[0m\n\x1b[0mAuthenticating against registry \x1b[32mDockerhub\x1b[39m\x1b[0m\nLogged in\n\x1b[0mUsing build variables commit \x1b[32mabc123\x1b[39m on branch \x1b[32mfeature1\x1b[39m\x1b[0m\n", out.String())
	assert.Equal(t, "error Code: 123 Message: build error\n", eout.String())
}

func TestBuild_BrokenOutput(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{BrokenOutput: true}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), out, eout, "Dockerfile", arrayFlags{}, false, false)

	assert.Equal(t, -7, code)
	assert.Equal(t, "\x1b[0mUsing CI \x1b[32mGitlab\x1b[39m\x1b[0m\n\x1b[0mUsing registry \x1b[32mDockerhub\x1b[39m\x1b[0m\n\x1b[0mAuthenticating against registry \x1b[32mDockerhub\x1b[39m\x1b[0m\nLogged in\n\x1b[0mUsing build variables commit \x1b[32mabc123\x1b[39m on branch \x1b[32mfeature1\x1b[39m\x1b[0m\n", out.String())
	assert.Equal(t, "unable to parse response: {\"code\":123,, Error: unexpected end of JSON input\n", eout.String())
}

func TestBuild_WithBuildArgs(t *testing.T) {
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("CI_COMMIT_SHA", "sha")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), out, eout, "Dockerfile", arrayFlags{"buildargs1=1", "buildargs2=2"}, false, false)
	assert.Equal(t, 0, code)

	assert.Equal(t, 4, len(client.BuildOptions[0].BuildArgs))
	assert.Equal(t, "1", *client.BuildOptions[0].BuildArgs["buildargs1"])
	assert.Equal(t, "2", *client.BuildOptions[0].BuildArgs["buildargs2"])
}

func TestBuild_WithStrangeBuildArg(t *testing.T) {
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("CI_COMMIT_SHA", "sha")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("buildargs4", "env-value")()
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), out, eout, "Dockerfile", arrayFlags{"buildargs1=1=1", "buildargs2", "buildargs3=", "buildargs4"}, false, false)
	assert.Equal(t, 0, code)

	assert.Equal(t, 4, len(client.BuildOptions[0].BuildArgs))
	assert.Equal(t, "1=1", *client.BuildOptions[0].BuildArgs["buildargs1"])
	assert.Equal(t, "env-value", *client.BuildOptions[0].BuildArgs["buildargs4"])
	assert.Equal(t, "", eout.String())
	assert.Equal(t, "\x1b[0mUsing CI \x1b[32mGitlab\x1b[39m\x1b[0m\n\x1b[0mUsing registry \x1b[32mDockerhub\x1b[39m\x1b[0m\n\x1b[0mAuthenticating against registry \x1b[32mDockerhub\x1b[39m\x1b[0m\nLogged in\n\x1b[0mUsing build variables commit \x1b[32msha\x1b[39m on branch \x1b[32mmaster\x1b[39m\x1b[0m\nignoring build-arg buildargs2\nignoring build-arg buildargs3\nBuild successful", out.String())
}

func TestBuild_WithSkipLogin(t *testing.T) {
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("CI_COMMIT_SHA", "sha")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), out, eout, "Dockerfile", arrayFlags{}, true, false)
	assert.Equal(t, 0, code)
	assert.Equal(t, "\x1b[0mUsing CI \x1b[32mGitlab\x1b[39m\x1b[0m\n\x1b[0mUsing registry \x1b[32mDockerhub\x1b[39m\x1b[0m\n\x1b[0mLogin \x1b[33mdisabled\x1b[39m\x1b[0m\n\x1b[0mUsing build variables commit \x1b[32msha\x1b[39m on branch \x1b[32mmaster\x1b[39m\x1b[0m\nBuild successful", out.String())
}

func TestBuild_FeatureBranch(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "feature1")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), out, eout, "Dockerfile", arrayFlags{}, false, false)

	assert.Equal(t, 0, code)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, 2, len(client.BuildOptions[0].BuildArgs))
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
	assert.Equal(t, "\x1b[0mUsing CI \x1b[32mGitlab\x1b[39m\x1b[0m\n\x1b[0mUsing registry \x1b[32mDockerhub\x1b[39m\x1b[0m\n\x1b[0mAuthenticating against registry \x1b[32mDockerhub\x1b[39m\x1b[0m\nLogged in\n\x1b[0mUsing build variables commit \x1b[32mabc123\x1b[39m on branch \x1b[32mfeature1\x1b[39m\x1b[0m\nBuild successful", out.String())
	assert.Equal(t, "", eout.String())
}

func TestBuild_DockerTagOverride(t *testing.T) {
	defer pkg.SetEnv("DOCKER_TAG", "override")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), out, eout, "Dockerfile", arrayFlags{}, false, false)

	assert.Equal(t, 0, code)
	assert.Equal(t, []string{"repo/build:override"}, client.BuildOptions[0].Tags)
	assert.Equal(t, []string{"repo/build:override"}, client.BuildOptions[0].CacheFrom)
	assert.Equal(t, "\x1b[0mUsing CI \x1b[32mnone\x1b[39m\x1b[0m\n\x1b[0mUsing registry \x1b[32mDockerhub\x1b[39m\x1b[0m\n\x1b[0mAuthenticating against registry \x1b[32mDockerhub\x1b[39m\x1b[0m\nLogged in\n\x1b[0mUsing build variables commit \x1b[32m\x1b[39m on branch \x1b[32m\x1b[39m\x1b[0m\noverriding docker tags with value from env DOCKER_TAG override\nBuild successful", out.String())
}

func TestBuild_MasterBranch(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	code := build(client, name, ioutil.NopCloser(buildContext), out, eout, "Dockerfile", arrayFlags{}, false, false)

	assert.Equal(t, 0, code)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[0].Tags)
	assert.Equal(t, "\x1b[0mUsing CI \x1b[32mGitlab\x1b[39m\x1b[0m\n\x1b[0mUsing registry \x1b[32mDockerhub\x1b[39m\x1b[0m\n\x1b[0mAuthenticating against registry \x1b[32mDockerhub\x1b[39m\x1b[0m\nLogged in\n\x1b[0mUsing build variables commit \x1b[32mabc123\x1b[39m on branch \x1b[32mmaster\x1b[39m\x1b[0m\nBuild successful", out.String())
	assert.Equal(t, "", eout.String())
}

func TestBuild_ParseError(t *testing.T) {
	response := `{"errorDetail":{"code":1,"message":"The command '/bin/sh -c yarn install  --frozen-lockfile' returned a non-zero code: 1"},"error":"The command '/bin/sh -c yarn install  --frozen-lockfile' returned a non-zero code: 1"}`
	r := &responsetype{}

	err := json.Unmarshal([]byte(response), &r)

	assert.NoError(t, err)
}

func TestBuild_BadDockerHost(t *testing.T) {
	defer pkg.SetEnv("DOCKER_HOST", "abc-123")()
	out := bytes.Buffer{}
	eout := bytes.Buffer{}
	exitCode := DoBuild(name, &out, &eout)
	assert.Equal(t, -1, exitCode)
	assert.Equal(t, "unable to parse docker host `abc-123`\n", out.String())
}

func TestBuild_UnreadableDockerignore(t *testing.T) {
	filename := filepath.Join(name, ".dockerignore")
	_ = os.Mkdir(filename, 0777)
	defer func() { _ = os.RemoveAll(filename) }()
	out := bytes.Buffer{}
	eout := bytes.Buffer{}
	exitCode := DoBuild(name, &out, &eout)
	assert.Equal(t, -2, exitCode)
	assert.Equal(t, fmt.Sprintf("read %s: is a directory\n", filename), out.String())
}

func TestBuild_Unreadable_Dockerfile(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{}

	code := build(client, name, ioutil.NopCloser(&brokenReader{}), out, eout, "Dockerfile", arrayFlags{}, false, false)

	assert.Equal(t, -5, code)
	assert.Equal(t, "read error\n", eout.String())
}

func TestBuild_HandleCaching(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
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

	buildContext, _ := archive.Generate("Dockerfile", dockerfile)
	code := build(client, name, ioutil.NopCloser(buildContext), out, eout, "Dockerfile", arrayFlags{}, false, false)

	assert.Equal(t, 0, code)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, 3, len(client.BuildOptions))
	assert.Equal(t, []string{"repo/reponame:build"}, client.BuildOptions[0].Tags)
	assert.Equal(t, []string{"repo/reponame:test"}, client.BuildOptions[1].Tags)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[2].Tags)
	assert.Equal(t, []string{"repo/reponame:build"}, client.BuildOptions[0].CacheFrom)
	assert.Equal(t, []string{"repo/reponame:test", "repo/reponame:build"}, client.BuildOptions[1].CacheFrom)
	assert.Equal(t, []string{"repo/reponame:master", "repo/reponame:latest", "repo/reponame:test", "repo/reponame:build"}, client.BuildOptions[2].CacheFrom)
	assert.Equal(t, "\x1b[0mUsing CI \x1b[32mGitlab\x1b[39m\x1b[0m\n\x1b[0mUsing registry \x1b[32mDockerhub\x1b[39m\x1b[0m\n\x1b[0mAuthenticating against registry \x1b[32mDockerhub\x1b[39m\x1b[0m\nLogged in\n\x1b[0mUsing build variables commit \x1b[32mabc123\x1b[39m on branch \x1b[32mmaster\x1b[39m\x1b[0m\nBuild successfulBuild successfulBuild successful", out.String())
	assert.Equal(t, "", eout.String())
}

func TestBuild_BrokenStage(t *testing.T) {
	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "reponame")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	defer pkg.SetEnv("DOCKERHUB_NAMESPACE", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
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

	buildContext, _ := archive.Generate("Dockerfile", dockerfile)
	code := build(client, name, ioutil.NopCloser(buildContext), out, eout, "Dockerfile", arrayFlags{}, false, false)

	assert.Equal(t, -7, code)
	assert.Equal(t, "build error\n", eout.String())
}

type brokenReader struct{}

func (b brokenReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

var _ io.Reader = &brokenReader{}
