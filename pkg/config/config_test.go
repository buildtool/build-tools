package config

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_AbsFail(t *testing.T) {
	os.Clearenv()

	abs = func(path string) (s string, e error) {
		return "", errors.New("abs-error")
	}

	_, err := Load("test")
	assert.EqualError(t, err, "abs-error")
	abs = filepath.Abs
}

func TestLoad_Empty(t *testing.T) {
	os.Clearenv()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)

	cfg, err := Load(name)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, "", cfg.CI.Selected)
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "", cfg.Registry.Selected)
}

func TestLoad_BrokenYAML(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	yaml := `ci: []
`
	_ = ioutil.WriteFile(filepath.Join(name, "buildtools.yaml"), []byte(yaml), 0777)

	cfg, err := Load(name)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig")
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, "", cfg.CI.Selected)
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "", cfg.Registry.Selected)
}

func TestLoad_UnreadableFile(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	filename := filepath.Join(name, "buildtools.yaml")
	_ = os.Mkdir(filename, 0777)

	cfg, err := Load(name)
	assert.EqualError(t, err, fmt.Sprintf("read %s: is a directory", filename))
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, "", cfg.CI.Selected)
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "", cfg.Registry.Selected)
}

func TestLoad_YAML(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	yaml := `
vcs:
  selected: gitlab
ci:
  selected: gitlab
registry:
  selected: quay
  dockerhub:
    repository: repo
    username: user
    password: pass
  ecr:
    url: 1234.ecr
    region: eu-west-1
  gitlab:
    repository: registry.gitlab.com/group/project
    token: token-value
  quay:
    repository: repo
    username: user
    password: pass
environments:
  - name: local
    context: docker-desktop
  - name: dev
    context: docker-desktop
    namespace: dev
`
	_ = ioutil.WriteFile(filepath.Join(name, "buildtools.yaml"), []byte(yaml), 0777)

	cfg, err := Load(name)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, "gitlab", cfg.CI.Selected)
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "quay", cfg.Registry.Selected)
	assert.Equal(t, &DockerhubRegistry{Repository: "repo", Username: "user", Password: "pass"}, cfg.Registry.Dockerhub)
	assert.Equal(t, &ECRRegistry{Url: "1234.ecr", Region: "eu-west-1"}, cfg.Registry.ECR)
	assert.Equal(t, &GitlabRegistry{Repository: "registry.gitlab.com/group/project", Token: "token-value"}, cfg.Registry.Gitlab)
	assert.Equal(t, &QuayRegistry{Repository: "repo", Username: "user", Password: "pass"}, cfg.Registry.Quay)
	assert.Equal(t, 2, len(cfg.Environments))
	assert.Equal(t, Environment{"local", "docker-desktop", ""}, cfg.Environments[0])
	devEnv := Environment{"dev", "docker-desktop", "dev"}
	assert.Equal(t, devEnv, cfg.Environments[1])

	currentEnv, err := cfg.CurrentEnvironment("dev")
	assert.NoError(t, err)
	assert.Equal(t, &devEnv, currentEnv)
	_, err = cfg.CurrentEnvironment("missing")
	assert.EqualError(t, err, "no environment matching missing found")
}

func TestLoad_YAML_DirStructure(t *testing.T) {
	os.Clearenv()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	yaml := `ci:
  selected: gitlab
registry:
  selected: quay
`
	_ = ioutil.WriteFile(filepath.Join(name, "buildtools.yaml"), []byte(yaml), 0777)
	subdir := "sub"
	_ = os.Mkdir(filepath.Join(name, subdir), 0777)
	yaml2 := `ci:
  selected: buildkite
`
	_ = ioutil.WriteFile(filepath.Join(name, subdir, "buildtools.yaml"), []byte(yaml2), 0777)

	cfg, err := Load(filepath.Join(name, subdir))
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, "buildkite", cfg.CI.Selected)
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "quay", cfg.Registry.Selected)
}

func TestLoad_ENV(t *testing.T) {
	_ = os.Setenv("CI", "gitlab")
	_ = os.Setenv("REGISTRY", "quay")
	cfg, err := Load(".")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, "gitlab", cfg.CI.Selected)
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "quay", cfg.Registry.Selected)
}
