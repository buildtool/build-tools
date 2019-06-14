package config

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGitlab_Identify(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_REGISTRY", "registry.gitlab.com")
	_ = os.Setenv("CI_REGISTRY_IMAGE", "registry.gitlab.com/group/image")
	_ = os.Setenv("CI_BUILD_TOKEN", "token")

	out := &bytes.Buffer{}
	cfg, err := Load(".", out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "registry.gitlab.com/group", registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestGitlab_RepositoryWithoutSlash(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_REGISTRY", "registry.gitlab.com")
	_ = os.Setenv("CI_REGISTRY_IMAGE", "registry.gitlab.com")
	_ = os.Setenv("CI_BUILD_TOKEN", "token")

	out := &bytes.Buffer{}
	cfg, err := Load(".", out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "registry.gitlab.com", registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestGitlab_RegistryFallback(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("REGISTRY", "gitlab")
	_ = os.Setenv("CI_REGISTRY", "registry.gitlab.com")
	_ = os.Setenv("CI_REGISTRY_IMAGE", "")
	_ = os.Setenv("CI_BUILD_TOKEN", "token")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)
	oldPwd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer os.Chdir(oldPwd)

	out := &bytes.Buffer{}
	cfg, err := Load(".", out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, fmt.Sprintf("registry.gitlab.com/%s", filepath.Base(dir)), registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestGitlab_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &GitlabRegistry{Registry: "registry.gitlab.com", Repository: "registry.gitlab.com/group/repo", Token: "token"}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.Nil(t, err)
	assert.Equal(t, "gitlab-ci-token", client.Username)
	assert.Equal(t, "token", client.Password)
	assert.Equal(t, "registry.gitlab.com", client.ServerAddress)
	assert.Equal(t, "Logged in\n", out.String())
}

func TestGitlab_LoginError(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &GitlabRegistry{}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.EqualError(t, err, "invalid username/password")
	assert.Equal(t, "", out.String())
}

func TestGitlab_GetAuthInfo(t *testing.T) {
	registry := &GitlabRegistry{Registry: "registry.gitlab.com", Repository: "registry.gitlab.com/group/repo", Token: "token"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6ImdpdGxhYi1jaS10b2tlbiIsInBhc3N3b3JkIjoidG9rZW4iLCJzZXJ2ZXJhZGRyZXNzIjoicmVnaXN0cnkuZ2l0bGFiLmNvbSJ9", auth)
}

func TestGitlab_Create(t *testing.T) {
	registry := &GitlabRegistry{}
	err := registry.Create("repo")
	assert.Nil(t, err)
}
