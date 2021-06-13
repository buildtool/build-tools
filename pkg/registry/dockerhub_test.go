package registry

import (
	"fmt"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg/docker"
)

func TestDockerhub_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &Dockerhub{Namespace: "repo", Username: "user", Password: "pass"}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.Nil(t, err)
	assert.Equal(t, "user", client.Username)
	assert.Equal(t, "pass", client.Password)
	assert.Equal(t, "", client.ServerAddress)
	logMock.Check(t, []string{"debug: Logged in\n"})
}

func TestDockerhub_LoginError(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &Dockerhub{}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.EqualError(t, err, "invalid username/password")
	logMock.Check(t, []string{"error: Unable to login\n"})
}

func TestDockerhub_GetAuthInfo(t *testing.T) {
	registry := &Dockerhub{Namespace: "repo", Username: "user", Password: "pass"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6InVzZXIiLCJwYXNzd29yZCI6InBhc3MifQ==", auth)
}

func TestDockerhub_Create(t *testing.T) {
	registry := &Dockerhub{}
	err := registry.Create("repo")
	assert.Nil(t, err)
}
