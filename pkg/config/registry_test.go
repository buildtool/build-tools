package config

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestDockerhub_Identify(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "repo", registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestDockerhub_Name(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("DOCKERHUB_REPOSITORY", "repo")
	_ = os.Setenv("DOCKERHUB_USERNAME", "user")
	_ = os.Setenv("DOCKERHUB_PASSWORD", "pass")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.Equal(t, "Dockerhub", registry.Name())
}

func TestEcr_Identify(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("ECR_URL", "url")
	_ = os.Setenv("ECR_REGION", "region")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "url", registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestEcr_Name(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("ECR_URL", "url")
	_ = os.Setenv("ECR_REGION", "region")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.Equal(t, "ECR", registry.Name())
}

func TestEcr_Identify_BrokenConfig(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("ECR_URL", "url")
	_ = os.Setenv("ECR_REGION", "region")
	_ = os.Setenv("AWS_CA_BUNDLE", "/missing/bundle")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.EqualError(t, err, "no Docker registry found")
	assert.Nil(t, registry)
	assert.Equal(t, "", out.String())
}

func TestGitlab_Identify(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_REGISTRY", "registry.gitlab.com")
	_ = os.Setenv("CI_REGISTRY_IMAGE", "registry.gitlab.com/group/image")
	_ = os.Setenv("CI_BUILD_TOKEN", "token")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "registry.gitlab.com/group", registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestGitlab_Name(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_REGISTRY", "registry.gitlab.com")
	_ = os.Setenv("CI_REGISTRY_IMAGE", "registry.gitlab.com/group/image")
	_ = os.Setenv("CI_BUILD_TOKEN", "token")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.Equal(t, "Gitlab", registry.Name())
}

func TestGitlab_RepositoryWithoutSlash(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_REGISTRY", "registry.gitlab.com")
	_ = os.Setenv("CI_REGISTRY_IMAGE", "registry.gitlab.com")
	_ = os.Setenv("CI_BUILD_TOKEN", "token")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "registry.gitlab.com", registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestGitlab_RegistryFallback(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_REGISTRY", "registry.gitlab.com")
	_ = os.Setenv("CI_REGISTRY_IMAGE", "")
	_ = os.Setenv("CI_BUILD_TOKEN", "token")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)
	oldPwd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer os.Chdir(oldPwd)

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "registry.gitlab.com", registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestQuay_Identify(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("QUAY_REPOSITORY", "repo")
	_ = os.Setenv("QUAY_USERNAME", "user")
	_ = os.Setenv("QUAY_PASSWORD", "pass")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "quay.io/repo", registry.RegistryUrl())
	assert.Equal(t, "", out.String())
}

func TestQuay_Name(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("QUAY_REPOSITORY", "repo")
	_ = os.Setenv("QUAY_USERNAME", "user")
	_ = os.Setenv("QUAY_PASSWORD", "pass")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.Equal(t, "Quay.io", registry.Name())
}
