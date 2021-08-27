package registry

import (
	"fmt"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg/docker"
)

func TestGithub_Name(t *testing.T) {
	registry := &Github{Repository: "repo", Username: "user", Password: "token"}

	assert.Equal(t, "Github", registry.Name())
}

func TestGithub_Configured(t *testing.T) {
	registry := &Github{Repository: "repo", Username: "user", Password: "token"}

	assert.True(t, registry.Configured())
}

func TestGithub_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &Github{Repository: "repo", Username: "user", Token: "token"}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.Nil(t, err)
	assert.Equal(t, "user", client.Username)
	assert.Equal(t, "token", client.Password)
	assert.Equal(t, "ghcr.io", client.ServerAddress)
	logMock.Check(t, []string{"debug: Logged in\n"})
}

func TestGithub_LoginError(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &Github{}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.EqualError(t, err, "invalid username/password")
	logMock.Check(t, []string{})
}

func TestGithub_GetAuthInfo(t *testing.T) {
	registry := &Github{Repository: "repo", Username: "user", Password: "token"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6InVzZXIiLCJwYXNzd29yZCI6InRva2VuIiwic2VydmVyYWRkcmVzcyI6ImdoY3IuaW8ifQ==", auth)
}

func TestGithub_Create(t *testing.T) {
	registry := &Github{}
	err := registry.Create("repo")
	assert.Nil(t, err)
}

func TestGithub_RegistryUrl(t *testing.T) {
	registry := &Github{Repository: "org/repo", Username: "user", Password: "token"}

	assert.Equal(t, "ghcr.io/org/repo", registry.RegistryUrl())
}
