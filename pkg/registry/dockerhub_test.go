package registry

import (
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

	registry := Identify()
	assert.NotNil(t, registry)
	assert.Equal(t, "repo", registry.RegistryUrl())
}

func TestDockerhub_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &dockerhub{repository: "repo", username: "user", password: "pass"}
	err := registry.Login(client)
	assert.Nil(t, err)
	assert.Equal(t, "user", client.Username)
	assert.Equal(t, "pass", client.Password)
	assert.Equal(t, "", client.ServerAddress)
}

func TestDockerhub_LoginError(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &dockerhub{}
	err := registry.Login(client)
	assert.EqualError(t, err, "invalid username/password")
}

func TestDockerhub_GetAuthInfo(t *testing.T) {
	registry := &dockerhub{repository: "repo", username: "user", password: "pass"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6InVzZXIiLCJwYXNzd29yZCI6InBhc3MifQ==", auth)
}

func TestDockerhub_Create(t *testing.T) {
	registry := &dockerhub{}
	err := registry.Create("repo")
	assert.Nil(t, err)
}
