package push

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"os"
	"testing"
)

func TestPush_NoCI(t *testing.T) {
	os.Clearenv()
	client := &docker.MockDocker{}

	err := Push(client, "Dockerfile")
	assert.NotNil(t, err)
	assert.EqualError(t, err, "no CI found")
}

func TestPush_NoRegistry(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")

	client := &docker.MockDocker{}
	err := Push(client, "Dockerfile")

	assert.NotNil(t, err)
	assert.EqualError(t, err, "no Docker registry found")
}

func TestPush_LoginFailure(t *testing.T) {
  os.Clearenv()
  _ = os.Setenv("GITLAB_CI", "1")
  _ = os.Setenv("CI_COMMIT_SHA", "abc123")
  _ = os.Setenv("CI_PROJECT_NAME", "reponame")
  _ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")
  _ = os.Setenv("ECR_URL", "ecr_url")

  client := &docker.MockDocker{}
  err := Push(client, "Dockerfile")

  assert.NotNil(t, err)
  assert.EqualError(t, err, "MissingRegion: could not find region configuration")
}

func TestPush_PushError(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	client := &docker.MockDocker{PushError: fmt.Errorf("unable to push layer")}
	err := Push(client, "Dockerfile")

	assert.NotNil(t, err)
	assert.EqualError(t, err, "unable to push layer")
}

func TestPush_PushFeatureBranch(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature1")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	client := &docker.MockDocker{}
	err := Push(client, "Dockerfile")

	assert.Nil(t, err)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:feature1"}, client.Images)
}

func TestPush_PushMasterBranch(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "master")
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	client := &docker.MockDocker{}
	err := Push(client, "Dockerfile")

	assert.Nil(t, err)
	assert.Equal(t, []string{"repo/reponame:abc123", "repo/reponame:master", "repo/reponame:latest"}, client.Images)
}
