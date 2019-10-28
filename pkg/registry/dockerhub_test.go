package registry

import (
	"bytes"
	"fmt"
	"github.com/sparetimecoders/build-tools/pkg/docker"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDockerhub_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &Dockerhub{Namespace: "repo", Username: "user", Password: "pass"}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.Nil(t, err)
	assert.Equal(t, "user", client.Username)
	assert.Equal(t, "pass", client.Password)
	assert.Equal(t, "", client.ServerAddress)
	assert.Equal(t, "Logged in\n", out.String())
}

func TestDockerhub_LoginError(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &Dockerhub{}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.EqualError(t, err, "invalid username/password")
	assert.Equal(t, "Unable to login\n", out.String())
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
