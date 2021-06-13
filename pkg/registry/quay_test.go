package registry

import (
	"fmt"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg/docker"
)

func TestQuay_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &Quay{Repository: "group", Username: "user", Password: "pass"}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)

	err := registry.Login(client)
	assert.Nil(t, err)
	assert.Equal(t, "user", client.Username)
	assert.Equal(t, "pass", client.Password)
	assert.Equal(t, "quay.io", client.ServerAddress)
	logMock.Check(t, []string{"debug: Logged in\n"})
}

func TestQuay_LoginError(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &Quay{}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)

	err := registry.Login(client)
	assert.EqualError(t, err, "invalid username/password")
	logMock.Check(t, []string{})
}

func TestQuay_GetAuthInfo(t *testing.T) {
	registry := &Quay{Repository: "repo", Username: "user", Password: "pass"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6InVzZXIiLCJwYXNzd29yZCI6InBhc3MiLCJzZXJ2ZXJhZGRyZXNzIjoicXVheS5pbyJ9", auth)
}

func TestQuay_Create(t *testing.T) {
	registry := &Quay{}
	err := registry.Create("repo")
	assert.Nil(t, err)
}
