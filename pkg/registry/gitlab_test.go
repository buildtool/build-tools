package registry

import (
	"fmt"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg/docker"
)

func TestGitlab_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &Gitlab{Registry: "registry.gitlab.com", Repository: "registry.gitlab.com/group/repo", Token: "token"}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.Nil(t, err)
	assert.Equal(t, "gitlab-ci-token", client.Username)
	assert.Equal(t, "token", client.Password)
	assert.Equal(t, "registry.gitlab.com", client.ServerAddress)
	logMock.Check(t, []string{"debug: Logged in\n"})
}

func TestGitlab_LoginError(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &Gitlab{}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.EqualError(t, err, "invalid username/password")
	logMock.Check(t, []string{})
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
