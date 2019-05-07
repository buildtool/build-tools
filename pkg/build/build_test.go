package build

import (
  "fmt"
  "github.com/docker/docker/pkg/archive"
  "github.com/stretchr/testify/assert"
  "gitlab.com/sparetimecoders/build-tools/pkg/docker"
  "io/ioutil"
  "os"
  "testing"
)

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
  _ = os.Setenv("GITLAB_CI", "1")
  _ = os.Setenv("CI_COMMIT_SHA", "abc123")
  _ = os.Setenv("CI_PROJECT_NAME", "reponame")
  _ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")

  client := &docker.MockDocker{}
  buildContext, _ := archive.Generate("Dockerfile", "FROM scratch")
  err := Build(client, ioutil.NopCloser(buildContext), "Dockerfile")

  assert.NotNil(t, err)
  assert.EqualError(t, err, "no Docker registry found")
}

func TestBuild_BuildError(t *testing.T) {
  os.Clearenv()
  _ = os.Setenv("GITLAB_CI", "1")
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
  assert.Error(t, err, "build error")
}

func TestBuild_FeatureBranch(t *testing.T) {
  os.Clearenv()
  _ = os.Setenv("GITLAB_CI", "1")
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
  _ = os.Setenv("GITLAB_CI", "1")
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
