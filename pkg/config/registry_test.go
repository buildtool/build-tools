package config

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg"
	"io/ioutil"
	"os"
	"testing"
)

func TestDockerhub_Identify(t *testing.T) {
	defer pkg.SetEnv("DOCKERHUB_REPOSITORY", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

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
	defer pkg.SetEnv("DOCKERHUB_REPOSITORY", "repo")()
	defer pkg.SetEnv("DOCKERHUB_USERNAME", "user")()
	defer pkg.SetEnv("DOCKERHUB_PASSWORD", "pass")()

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.Equal(t, "Dockerhub", registry.Name())
}

func TestEcr_Identify(t *testing.T) {
	defer pkg.SetEnv("ECR_URL", "url")()
	defer pkg.SetEnv("ECR_REGION", "region")()

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
	defer pkg.SetEnv("ECR_URL", "url")()
	defer pkg.SetEnv("ECR_REGION", "region")()

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.Equal(t, "ECR", registry.Name())
}

func TestEcr_Identify_BrokenConfig(t *testing.T) {
	defer pkg.SetEnv("ECR_URL", "url")()
	defer pkg.SetEnv("ECR_REGION", "region")()
	defer pkg.SetEnv("AWS_CA_BUNDLE", "/missing/bundle")()

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.EqualError(t, err, "no Docker registry found")
	assert.Nil(t, registry)
	assert.Equal(t, "", out.String())
}

func TestGitlab_Identify(t *testing.T) {
	defer pkg.SetEnv("CI_REGISTRY", "registry.gitlab.com")()
	defer pkg.SetEnv("CI_REGISTRY_IMAGE", "registry.gitlab.com/group/image")()
	defer pkg.SetEnv("CI_BUILD_TOKEN", "token")()

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
	defer pkg.SetEnv("CI_REGISTRY", "registry.gitlab.com")()
	defer pkg.SetEnv("CI_REGISTRY_IMAGE", "registry.gitlab.com/group/image")()
	defer pkg.SetEnv("CI_BUILD_TOKEN", "token")()

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.Equal(t, "Gitlab", registry.Name())
}

func TestGitlab_RepositoryWithoutSlash(t *testing.T) {
	defer pkg.SetEnv("CI_REGISTRY", "registry.gitlab.com")()
	defer pkg.SetEnv("CI_REGISTRY_IMAGE", "registry.gitlab.com")()
	defer pkg.SetEnv("CI_BUILD_TOKEN", "token")()

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
	defer pkg.SetEnv("CI_REGISTRY", "registry.gitlab.com")()
	defer pkg.SetEnv("CI_REGISTRY_IMAGE", "")()
	defer pkg.SetEnv("CI_BUILD_TOKEN", "token")()

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
	defer pkg.SetEnv("QUAY_REPOSITORY", "repo")()
	defer pkg.SetEnv("QUAY_USERNAME", "user")()
	defer pkg.SetEnv("QUAY_PASSWORD", "pass")()

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
	defer pkg.SetEnv("QUAY_REPOSITORY", "repo")()
	defer pkg.SetEnv("QUAY_USERNAME", "user")()
	defer pkg.SetEnv("QUAY_PASSWORD", "pass")()

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.Equal(t, "Quay.io", registry.Name())
}
