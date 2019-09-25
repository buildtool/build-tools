package registry

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"testing"
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
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.Nil(t, err)
	assert.Equal(t, "user", client.Username)
	assert.Equal(t, "token", client.Password)
	assert.Equal(t, "docker.pkg.github.com", client.ServerAddress)
	assert.Equal(t, "Logged in\n", out.String())
}

func TestGithub_LoginError(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &Github{}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.EqualError(t, err, "invalid username/password")
	assert.Equal(t, "", out.String())
}

func TestGithub_GetAuthInfo(t *testing.T) {
	registry := &Github{Repository: "repo", Username: "user", Password: "token"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6InVzZXIiLCJwYXNzd29yZCI6InRva2VuIiwic2VydmVyYWRkcmVzcyI6ImRvY2tlci5wa2cuZ2l0aHViLmNvbSJ9", auth)
}

func TestGithub_Create(t *testing.T) {
	registry := &Github{}
	err := registry.Create("repo")
	assert.Nil(t, err)
}

func TestGithub_RegistryUrl(t *testing.T) {
	registry := &Github{Repository: "org/repo", Username: "user", Password: "token"}

	assert.Equal(t, "docker.pkg.github.com/org/repo", registry.RegistryUrl())
}
