package config

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sparetimecoders/build-tools/pkg"
	"github.com/sparetimecoders/build-tools/pkg/ci"
	"github.com/sparetimecoders/build-tools/pkg/registry"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var name string

func TestMain(m *testing.M) {
	tempDir := setup()
	code := m.Run()
	teardown(tempDir)
	os.Exit(code)
}

func setup() string {
	name, _ = ioutil.TempDir(os.TempDir(), "build-tools")

	return name
}

func teardown(tempDir string) {
	_ = os.RemoveAll(tempDir)
}

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
	defer func() { _ = os.RemoveAll(name) }()

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, ci.No{}.Name(), cfg.CurrentCI().Name())
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "", cfg.Scaffold.RegistryUrl)
	assert.Equal(t, "", out.String())
}

func TestLoad_BrokenYAML(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `ci: []
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig")
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, ci.No{}.Name(), cfg.CurrentCI().Name())
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "", cfg.Scaffold.RegistryUrl)
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s/.buildtools.yaml'\x1b[39m\x1b[0m\n", name), out.String())
}

func TestLoad_UnreadableFile(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	filename := filepath.Join(name, ".buildtools.yaml")
	_ = os.Mkdir(filename, 0777)

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.EqualError(t, err, fmt.Sprintf("read %s: is a directory", filename))
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, ci.No{}.Name(), cfg.CurrentCI().Name())
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "", cfg.Scaffold.RegistryUrl)
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s/.buildtools.yaml'\x1b[39m\x1b[0m\n", name), out.String())
}

func TestLoad_YAML(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
scaffold:
  registry: quay.io
registry:
  ecr:
    url: 1234.ecr
    region: eu-west-1
environments:
  local:
    context: docker-desktop
  dev:
    context: docker-desktop
    namespace: dev
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, ci.No{}.Name(), cfg.CurrentCI().Name())

	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "quay.io", cfg.Scaffold.RegistryUrl)
	assert.Equal(t, "eu-west-1", cfg.CurrentRegistry().(*registry.ECR).Region)
	assert.Equal(t, "1234.ecr", cfg.CurrentRegistry().(*registry.ECR).Url)
	assert.Equal(t, 2, len(cfg.Environments))
	assert.Equal(t, Environment{Context: "docker-desktop"}, cfg.Environments["local"])
	devEnv := Environment{Context: "docker-desktop", Namespace: "dev"}
	assert.Equal(t, devEnv, cfg.Environments["dev"])

	currentEnv, err := cfg.CurrentEnvironment("dev")
	assert.NoError(t, err)
	assert.Equal(t, &devEnv, currentEnv)
	_, err = cfg.CurrentEnvironment("missing")
	assert.EqualError(t, err, "no environment matching missing found")
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s/.buildtools.yaml'\x1b[39m\x1b[0m\n", name), out.String())
}

func TestLoad_BrokenYAML_From_Env(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `ci: []
`
	defer pkg.SetEnv("BUILDTOOLS_CONTENT", yaml)()

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig")
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, ci.No{}.Name(), cfg.CurrentCI().Name())
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "", cfg.Scaffold.RegistryUrl)
	assert.Equal(t, "Parsing config from env: BUILDTOOLS_CONTENT\n", out.String())
}

func TestLoad_YAML_From_Env(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
environments:
  local:
    context: docker-desktop
  dev:
    context: docker-desktop
    namespace: dev
`
	defer pkg.SetEnv("BUILDTOOLS_CONTENT", yaml)()

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 2, len(cfg.Environments))
	assert.Equal(t, Environment{Context: "docker-desktop"}, cfg.Environments["local"])
	devEnv := Environment{Context: "docker-desktop", Namespace: "dev"}
	assert.Equal(t, devEnv, cfg.Environments["dev"])

	currentEnv, err := cfg.CurrentEnvironment("dev")
	assert.NoError(t, err)
	assert.Equal(t, &devEnv, currentEnv)
	_, err = cfg.CurrentEnvironment("missing")
	assert.EqualError(t, err, "no environment matching missing found")
	assert.Equal(t, "Parsing config from env: BUILDTOOLS_CONTENT\n", out.String())
}

func TestLoad_YAML_DirStructure(t *testing.T) {
	os.Clearenv()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `scaffold:
  ci:
  registry: quay.io
registry:
  dockerhub:
    repository: test
environments:
  test:
    context: abc
  local:
    context: def
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)
	subdir := "sub"
	_ = os.Mkdir(filepath.Join(name, subdir), 0777)
	yaml2 := `scaffold:
  ci:
environments:
  test:
    context: ghi
`
	_ = ioutil.WriteFile(filepath.Join(name, subdir, ".buildtools.yaml"), []byte(yaml2), 0777)

	out := &bytes.Buffer{}
	cfg, err := Load(filepath.Join(name, subdir), out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, ci.No{}.Name(), cfg.CurrentCI().Name())
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "quay.io", cfg.Scaffold.RegistryUrl)
	currentRegistry := cfg.CurrentRegistry()
	assert.NotNil(t, currentRegistry)
	assert.Equal(t, 2, len(cfg.Environments))
	assert.Equal(t, "ghi", cfg.Environments["test"].Context)
	assert.Equal(t, "def", cfg.Environments["local"].Context)
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s/sub/.buildtools.yaml'\x1b[39m\x1b[0m\n\x1b[0mMerging with config from file: \x1b[32m'%s/.buildtools.yaml'\x1b[39m\x1b[0m\n", name, name), out.String())
}

func TestLoad_YAML_Multiple_Registry(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
registry:
  ecr:
    url: 1234.ecr
    region: eu-west-1
  dockerhub:
    repository: dockerhub
environments:
  local:
    context: docker-desktop
  dev:
    context: docker-desktop
    namespace: dev
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	out := &bytes.Buffer{}
	_, err := Load(name, out)
	assert.EqualError(t, err, "registry alread defined, please check configuration")
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s/.buildtools.yaml'\x1b[39m\x1b[0m\n", name), out.String())
}

func TestLoad_YAML_Scaffold_Multiple_CI(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
scaffold:
  ci: 
    gitlab:
      group: group
      token: token
    buildkite:
      token: token
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	out := &bytes.Buffer{}
	_, err := Load(name, out)
	assert.EqualError(t, err, "scaffold CI already defined, please check configuration")
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s/.buildtools.yaml'\x1b[39m\x1b[0m\n", name), out.String())
}

func TestLoad_YAML_Scaffold_Multiple_VCS(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
scaffold:
  vcs: 
    gitlab:
      group: group
    github:
      token: token
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	out := &bytes.Buffer{}
	_, err := Load(name, out)
	assert.EqualError(t, err, "scaffold VCS already defined, please check configuration")
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s/.buildtools.yaml'\x1b[39m\x1b[0m\n", name), out.String())
}
