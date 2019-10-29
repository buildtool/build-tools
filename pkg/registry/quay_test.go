package registry

import (
	"bytes"
	"fmt"
	"github.com/buildtool/build-tools/pkg/docker"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQuay_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &Quay{Repository: "group", Username: "user", Password: "pass"}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.Nil(t, err)
	assert.Equal(t, "user", client.Username)
	assert.Equal(t, "pass", client.Password)
	assert.Equal(t, "quay.io", client.ServerAddress)
	assert.Equal(t, "Logged in\n", out.String())
}

func TestQuay_LoginError(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &Quay{}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.EqualError(t, err, "invalid username/password")
	assert.Equal(t, "", out.String())
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
