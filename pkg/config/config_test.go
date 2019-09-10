package config

import (
	"bytes"
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

	out := &bytes.Buffer{}
	_, err := Load("test", out)
	assert.EqualError(t, err, "abs-error")
	assert.Equal(t, "", out.String())
	abs = filepath.Abs
}

func TestLoad_Empty(t *testing.T) {
	os.Clearenv()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, "", cfg.CI.Selected)
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "", cfg.Registry.Selected)
	assert.Equal(t, "", out.String())
}

func TestLoad_BrokenYAML(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	yaml := `ci: []
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig")
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, "", cfg.CI.Selected)
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "", cfg.Registry.Selected)
	assert.Equal(t, fmt.Sprintf("Parsing config from file: '%s/.buildtools.yaml'\n", name), out.String())
}

func TestLoad_UnreadableFile(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	filename := filepath.Join(name, ".buildtools.yaml")
	_ = os.Mkdir(filename, 0777)

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.EqualError(t, err, fmt.Sprintf("read %s: is a directory", filename))
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, "", cfg.CI.Selected)
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "", cfg.Registry.Selected)
	assert.Equal(t, fmt.Sprintf("Parsing config from file: '%s/.buildtools.yaml'\n", name), out.String())
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
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
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
	assert.Equal(t, fmt.Sprintf("Parsing config from file: '%s/.buildtools.yaml'\n", name), out.String())
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
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)
	subdir := "sub"
	_ = os.Mkdir(filepath.Join(name, subdir), 0777)
	yaml2 := `ci:
  selected: buildkite
`
	_ = ioutil.WriteFile(filepath.Join(name, subdir, ".buildtools.yaml"), []byte(yaml2), 0777)

	out := &bytes.Buffer{}
	cfg, err := Load(filepath.Join(name, subdir), out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, "buildkite", cfg.CI.Selected)
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "quay", cfg.Registry.Selected)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, fmt.Sprintf("Parsing config from file: '%s/.buildtools.yaml'\nParsing config from file: '%s/sub/.buildtools.yaml'\n", name, name), out.String())
}

func TestLoad_ENV(t *testing.T) {
	_ = os.Setenv("CI", "gitlab")
	_ = os.Setenv("REGISTRY", "quay")
	out := &bytes.Buffer{}
	cfg, err := Load(".", out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, "gitlab", cfg.CI.Selected)
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "quay", cfg.Registry.Selected)
	assert.Equal(t, "", out.String())
}

func TestLoad_Selected_VCS_Azure(t *testing.T) {
	_ = os.Setenv("CI", "azure")
	_ = os.Setenv("VCS", "azure")
	_ = os.Setenv("REGISTRY", "dockerhub")
	out := &bytes.Buffer{}
	cfg, err := Load(".", out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.VCS)
	assert.Equal(t, "azure", cfg.VCS.Selected)
	vcs := cfg.CurrentVCS()
	assert.Equal(t, "Azure", vcs.Name())
	assert.Equal(t, "", out.String())
}

func TestLoad_Selected_VCS_Github(t *testing.T) {
	_ = os.Setenv("CI", "buildkite")
	_ = os.Setenv("VCS", "github")
	_ = os.Setenv("REGISTRY", "dockerhub")
	out := &bytes.Buffer{}
	cfg, err := Load(".", out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.VCS)
	assert.Equal(t, "github", cfg.VCS.Selected)
	vcs := cfg.CurrentVCS()
	assert.Equal(t, "Github", vcs.Name())
	assert.Equal(t, "", out.String())
}

func TestLoad_Selected_VCS_Gitlab(t *testing.T) {
	_ = os.Setenv("CI", "gitlab")
	_ = os.Setenv("VCS", "gitlab")
	_ = os.Setenv("REGISTRY", "dockerhub")
	out := &bytes.Buffer{}
	cfg, err := Load(".", out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.VCS)
	assert.Equal(t, "gitlab", cfg.VCS.Selected)
	vcs := cfg.CurrentVCS()
	assert.Equal(t, "Gitlab", vcs.Name())
	assert.Equal(t, "", out.String())
}

func TestLoad_Selected_Registry_Dockerhub(t *testing.T) {
	_ = os.Setenv("CI", "gitlab")
	_ = os.Setenv("REGISTRY", "dockerhub")
	out := &bytes.Buffer{}
	cfg, err := Load(".", out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, "gitlab", cfg.CI.Selected)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "dockerhub", cfg.Registry.Selected)
	assert.Equal(t, "", out.String())
}

func TestLoad_Selected_Registry_ECR(t *testing.T) {
	_ = os.Setenv("CI", "gitlab")
	_ = os.Setenv("REGISTRY", "ecr")
	out := &bytes.Buffer{}
	cfg, err := Load(".", out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, "gitlab", cfg.CI.Selected)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "ecr", cfg.Registry.Selected)
	assert.Equal(t, "", out.String())
}

func TestLoad_Selected_Registry_Gitlab(t *testing.T) {
	_ = os.Setenv("CI", "gitlab")
	_ = os.Setenv("REGISTRY", "gitlab")
	out := &bytes.Buffer{}
	cfg, err := Load(".", out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, "gitlab", cfg.CI.Selected)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "gitlab", cfg.Registry.Selected)
	assert.Equal(t, "", out.String())
}

func TestLoad_Selected_Registry_Quay(t *testing.T) {
	_ = os.Setenv("CI", "gitlab")
	_ = os.Setenv("REGISTRY", "quay")
	out := &bytes.Buffer{}
	cfg, err := Load(".", out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, "gitlab", cfg.CI.Selected)
	registry, err := cfg.CurrentRegistry()
	assert.NoError(t, err)
	assert.NotNil(t, registry)
	assert.Equal(t, "quay", cfg.Registry.Selected)
	assert.Equal(t, "", out.String())
}
