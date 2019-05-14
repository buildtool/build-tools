package build

import (
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
	oldPwd, tempDir := setup()
	code := m.Run()
	teardown(oldPwd, tempDir)
	os.Exit(code)
}

func setup() (string, string) {
	oldPwd, _ := os.Getwd()
	name, _ = ioutil.TempDir(os.TempDir(), "build-tools")
	_ = os.Chdir(name)
	os.Clearenv()

	return oldPwd, name
}

func teardown(oldPwd, tempDir string) {
	_ = os.RemoveAll(tempDir)
	_ = os.Chdir(oldPwd)
}

func TestBuild_BrokenConfig(t *testing.T) {
	yaml := `ci: [] `
	file := filepath.Join(name, ".buildtools.yaml")
	_ = ioutil.WriteFile(file, []byte(yaml), 0777)
	defer os.Remove(file)

	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	err := Build(client, ioutil.NopCloser(buildContext), "Dockerfile")

	assert.NotNil(t, err)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig")
}

func TestBuild_NoCI(t *testing.T) {
	os.Clearenv()
	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	err := Build(client, ioutil.NopCloser(buildContext), "Dockerfile")

	assert.NotNil(t, err)
	assert.EqualError(t, err, "no CI found")
}

func TestBuild_NoRegistry(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")

	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	err := Build(client, ioutil.NopCloser(buildContext), "Dockerfile")

	assert.NotNil(t, err)
	assert.EqualError(t, err, "no Docker registry found")
}

func TestBuild_LoginError(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	err := Build(client, ioutil.NopCloser(buildContext), "Dockerfile")

	assert.NotNil(t, err)
	assert.EqualError(t, err, "invalid username/password")
}

func TestBuild_BuildError(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	client := &docker.MockDocker{BuildError: fmt.Errorf("build error")}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	err := Build(client, ioutil.NopCloser(buildContext), "Dockerfile")

	assert.NotNil(t, err)
	assert.EqualError(t, err, "build error")
}

func TestBuild_FeatureBranch(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	err := Build(client, ioutil.NopCloser(buildContext), "Dockerfile")

	assert.Nil(t, err)
	assert.Equal(t, "Dockerfile", client.BuildOptions.Dockerfile)
	assert.Equal(t, int64(3*1024*1024*1024), client.BuildOptions.Memory)
	assert.Equal(t, int64(-1), client.BuildOptions.MemorySwap)
	assert.Equal(t, true, client.BuildOptions.Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions.ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:feature1"}, client.BuildOptions.Tags)
}

func TestBuild_MasterBranch(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "master")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	client := &docker.MockDocker{}
	buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
	err := Build(client, ioutil.NopCloser(buildContext), "Dockerfile")

	assert.Nil(t, err)
	assert.Equal(t, "Dockerfile", client.BuildOptions.Dockerfile)
	assert.Equal(t, int64(3*1024*1024*1024), client.BuildOptions.Memory)
	assert.Equal(t, int64(-1), client.BuildOptions.MemorySwap)
	assert.Equal(t, true, client.BuildOptions.Remove)
	assert.Equal(t, int64(256*1024*1024), client.BuildOptions.ShmSize)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.BuildOptions.Tags)
}
