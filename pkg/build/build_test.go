package build

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/pkg/archive"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
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
	err := build(client, name, ioutil.NopCloser(buildContext), "Dockerfile", out, eout)

	absPath, _ := filepath.Abs(filepath.Join(name, ".buildtools.yaml"))
	assert.NotNil(t, err)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig")
	assert.Equal(t, fmt.Sprintf("Parsing config from file: '%s'\n", absPath), out.String())
	assert.Equal(t, "", eout.String())
}

func TestBuild_NoRegistry(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	err := build(client, name, ioutil.NopCloser(buildContext), "Dockerfile", out, eout)

	assert.NotNil(t, err)
	assert.EqualError(t, err, "no Docker registry found")
	assert.Equal(t, "", out.String())
	assert.Equal(t, "", eout.String())
}

func TestBuild_LoginError(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	err := build(client, name, ioutil.NopCloser(buildContext), "Dockerfile", out, eout)

	assert.NotNil(t, err)
	assert.EqualError(t, err, "invalid username/password")
	assert.Equal(t, "Unable to login\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestBuild_BuildError(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{BuildError: []error{fmt.Errorf("build error")}}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	err := build(client, name, ioutil.NopCloser(buildContext), "Dockerfile", out, eout)

	assert.NotNil(t, err)
	assert.EqualError(t, err, "build error")
	assert.Equal(t, "Logged in\n", out.String())
	assert.Equal(t, "", eout.String())
}

func TestBuild_BuildResponseError(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{ResponseError: fmt.Errorf("build error")}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	err := build(client, name, ioutil.NopCloser(buildContext), "Dockerfile", out, eout)

	assert.NotNil(t, err)
	assert.EqualError(t, err, "build error")
	assert.Equal(t, "Logged in\n", out.String())
	assert.Equal(t, "Code: 123 Message: build error\n", eout.String())
}

func TestBuild_BrokenOutput(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{BrokenOutput: true}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	err := build(client, name, ioutil.NopCloser(buildContext), "Dockerfile", out, eout)

	assert.NotNil(t, err)
	assert.EqualError(t, err, "unexpected end of JSON input")
	assert.Equal(t, "Logged in\n", out.String())
	assert.Equal(t, "Unable to parse response: {\"code\":123,, Error: unexpected end of JSON input\n", eout.String())
}

func TestBuild_FeatureBranch(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	err := build(client, name, ioutil.NopCloser(buildContext), "Dockerfile", out, eout)

	assert.Nil(t, err)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, 2, len(client.BuildOptions[0].BuildArgs))
	assert.Equal(t, "abc123", *client.BuildOptions[0].BuildArgs["CI_COMMIT"])
	assert.Equal(t, "feature1", *client.BuildOptions[0].BuildArgs["CI_BRANCH"])
	assert.Equal(t, int64(3*1024*1024*1024), client.BuildOptions[0].Memory)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:feature1"}, client.BuildOptions[0].Tags)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(3*1024*1024*1024), client.BuildOptions[0].Memory)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:feature1"}, client.BuildOptions[0].Tags)
	assert.Equal(t, "Logged in\nBuild successful", out.String())
	assert.Equal(t, "", eout.String())
}

func TestBuild_MasterBranch(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "master")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	err := build(client, name, ioutil.NopCloser(buildContext), "Dockerfile", out, eout)

	assert.Nil(t, err)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(3*1024*1024*1024), client.BuildOptions[0].Memory)
	assert.Equal(t, int64(-1), client.BuildOptions[0].MemorySwap)
	assert.Equal(t, true, client.BuildOptions[0].Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions[0].ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions[0].Tags)
	assert.Equal(t, "Logged in\nBuild successful", out.String())
	assert.Equal(t, "", eout.String())
}

func TestBuild_ParseError(t *testing.T) {
	response := `{"errorDetail":{"code":1,"message":"The command '/bin/sh -c yarn install  --frozen-lockfile' returned a non-zero code: 1"},"error":"The command '/bin/sh -c yarn install  --frozen-lockfile' returned a non-zero code: 1"}`
	r := &responsetype{}

	err := json.Unmarshal([]byte(response), &r)

	assert.NoError(t, err)
}

func TestBuild_BadDockerHost(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("DOCKER_HOST", "abc-123")
	out := bytes.Buffer{}
	eout := bytes.Buffer{}
	exitCode := 0
	DoBuild(name, &out, &eout, func(code int) {
		exitCode = code
	})
	assert.Equal(t, -1, exitCode)
	assert.Equal(t, "unable to parse docker host `abc-123`\n", out.String())
}

func TestBuild_UnreadableDockerignore(t *testing.T) {
	filename := filepath.Join(name, ".dockerignore")
	_ = os.Mkdir(filename, 0777)
	os.Clearenv()
	out := bytes.Buffer{}
	eout := bytes.Buffer{}
	exitCode := 0
	DoBuild(name, &out, &eout, func(code int) {
		exitCode = code
	})
	assert.Equal(t, -2, exitCode)
	assert.Equal(t, fmt.Sprintf("read %s: is a directory\n", filename), out.String())
}

func TestBuild_Unreadable_Dockerfile(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "master")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{}

	err := build(client, name, ioutil.NopCloser(&brokenReader{}), "Dockerfile", out, eout)

	assert.EqualError(t, err, "read error")
}

func TestBuild_HandleCaching(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "master")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

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
	err := build(client, name, ioutil.NopCloser(buildContext), "Dockerfile", out, eout)

	assert.Nil(t, err)
	assert.Equal(t, "Dockerfile", client.BuildOptions[0].Dockerfile)
	assert.Equal(t, int64(3*1024*1024*1024), client.BuildOptions[0].Memory)
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
	assert.Equal(t, "Logged in\nBuild successfulBuild successfulBuild successful", out.String())
	assert.Equal(t, "", eout.String())
}

func TestBuild_BrokenStage(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "master")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

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
	err := build(client, name, ioutil.NopCloser(buildContext), "Dockerfile", out, eout)

	assert.EqualError(t, err, "build error")
}

type brokenReader struct{}

func (b brokenReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

var _ io.Reader = &brokenReader{}
