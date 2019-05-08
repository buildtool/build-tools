package registry

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"os"
	"testing"
)

func TestGitlab_Identify(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_REGISTRY_IMAGE", "image")
	_ = os.Setenv("CI_REGISTRY", "registry.gitlab.com/group/image")
	_ = os.Setenv("CI_BUILD_TOKEN", "token")

	registry := Identify()
	assert.NotNil(t, registry)
	assert.Equal(t, "registry.gitlab.com/group", registry.RegistryUrl())
}

func TestGitlab_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &gitlab{url: "registry.gitlab.com/group/repo", token: "token"}
	err := registry.Login(client)
	assert.Nil(t, err)
	assert.Equal(t, "gitlab-ci-token", client.Username)
	assert.Equal(t, "token", client.Password)
	assert.Equal(t, "registry.gitlab.com/group/repo", client.ServerAddress)
}

func TestGitlab_LoginError(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &gitlab{}
	err := registry.Login(client)
	assert.EqualError(t, err, "invalid username/password")
}

func TestGitlab_GetAuthInfo(t *testing.T) {
	registry := &gitlab{url: "registry.gitlab.com/group/repo", token: "token"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6ImdpdGxhYi1jaS10b2tlbiIsInBhc3N3b3JkIjoidG9rZW4iLCJzZXJ2ZXJhZGRyZXNzIjoicmVnaXN0cnkuZ2l0bGFiLmNvbS9ncm91cC9yZXBvIn0=", auth)
}

func TestGitlab_Create(t *testing.T) {
	registry := &gitlab{}
	err := registry.Create("repo")
	assert.Nil(t, err)
}
