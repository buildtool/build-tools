package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"os"
	"testing"
)

func TestQuay_Identify(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("QUAY_REPOSITORY", "repo")
	_ = os.Setenv("QUAY_USERNAME", "user")
	_ = os.Setenv("QUAY_PASSWORD", "pass")

	cfg, err := Load(".")
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "quay.io/repo", registry.RegistryUrl())
}

func TestQuay_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &QuayRegistry{Repository: "group", Username: "user", Password: "pass"}
	err := registry.Login(client)
	assert.Nil(t, err)
	assert.Equal(t, "user", client.Username)
	assert.Equal(t, "pass", client.Password)
	assert.Equal(t, "quay.io", client.ServerAddress)
}

func TestQuay_LoginError(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &QuayRegistry{}
	err := registry.Login(client)
	assert.EqualError(t, err, "invalid username/password")
}

func TestQuay_GetAuthInfo(t *testing.T) {
	registry := &QuayRegistry{Repository: "repo", Username: "user", Password: "pass"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6InVzZXIiLCJwYXNzd29yZCI6InBhc3MiLCJzZXJ2ZXJhZGRyZXNzIjoicXVheS5pbyJ9", auth)
}

func TestQuay_Create(t *testing.T) {
	registry := &QuayRegistry{}
	err := registry.Create("repo")
	assert.Nil(t, err)
}
