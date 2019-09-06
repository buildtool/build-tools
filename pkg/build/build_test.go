package build

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/pkg/archive"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
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
	defer os.Remove(file)

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
	client := &docker.MockDocker{BuildError: fmt.Errorf("build error")}
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
	assert.Equal(t, "Dockerfile", client.BuildOptions.Dockerfile)
	assert.Equal(t, 2, len(client.BuildOptions.BuildArgs))
	assert.Equal(t, "abc123", *client.BuildOptions.BuildArgs["CI_COMMIT"])
	assert.Equal(t, "feature1", *client.BuildOptions.BuildArgs["CI_BRANCH"])
	assert.Equal(t, int64(3*1024*1024*1024), client.BuildOptions.Memory)
	assert.Equal(t, int64(-1), client.BuildOptions.MemorySwap)
	assert.Equal(t, true, client.BuildOptions.Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions.ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:feature1"}, client.BuildOptions.Tags)
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
	assert.Equal(t, "Dockerfile", client.BuildOptions.Dockerfile)
	assert.Equal(t, int64(3*1024*1024*1024), client.BuildOptions.Memory)
	assert.Equal(t, int64(-1), client.BuildOptions.MemorySwap)
	assert.Equal(t, true, client.BuildOptions.Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions.ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions.Tags)
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
