package registry

import (
	"bytes"
	"fmt"
	"github.com/buildtool/build-tools/pkg/docker"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGitlab_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &Gitlab{Registry: "registry.gitlab.com", Repository: "registry.gitlab.com/group/repo", Token: "token"}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.Nil(t, err)
	assert.Equal(t, "gitlab-ci-token", client.Username)
	assert.Equal(t, "token", client.Password)
	assert.Equal(t, "registry.gitlab.com", client.ServerAddress)
	assert.Equal(t, "Logged in\n", out.String())
}

func TestGitlab_LoginError(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &Gitlab{}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.EqualError(t, err, "invalid username/password")
	assert.Equal(t, "", out.String())
}

func TestGitlab_GetAuthInfo(t *testing.T) {
	registry := &Gitlab{Registry: "registry.gitlab.com", Repository: "registry.gitlab.com/group/repo", Token: "token"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6ImdpdGxhYi1jaS10b2tlbiIsInBhc3N3b3JkIjoidG9rZW4iLCJzZXJ2ZXJhZGRyZXNzIjoicmVnaXN0cnkuZ2l0bGFiLmNvbSJ9", auth)
}

func TestGitlab_Create(t *testing.T) {
	registry := &Gitlab{}
	err := registry.Create("repo")
	assert.Nil(t, err)
}
