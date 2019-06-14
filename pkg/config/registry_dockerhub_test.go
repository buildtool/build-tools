package config

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"os"
	"testing"
)

func TestDockerhub_Identify(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	out := &bytes.Buffer{}
	cfg, err := Load(".", out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "repo", registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestDockerhub_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &DockerhubRegistry{Repository: "repo", Username: "user", Password: "pass"}
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
	registry := &DockerhubRegistry{}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.EqualError(t, err, "invalid username/password")
	assert.Equal(t, "Unable to login\n", out.String())
}

func TestDockerhub_GetAuthInfo(t *testing.T) {
	registry := &DockerhubRegistry{Repository: "repo", Username: "user", Password: "pass"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6InVzZXIiLCJwYXNzd29yZCI6InBhc3MifQ==", auth)
}

func TestDockerhub_Create(t *testing.T) {
	registry := &DockerhubRegistry{}
	err := registry.Create("repo")
	assert.Nil(t, err)
}
