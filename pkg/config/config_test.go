package config

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/buildtool/build-tools/pkg"
	"github.com/buildtool/build-tools/pkg/ci"
	"github.com/buildtool/build-tools/pkg/registry"
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
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s/.buildtools.yaml'\x1b[39m\x1b[0m\n", name), out.String())
}

func TestLoad_YAML(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
registry:
  ecr:
    url: 1234.dkr.ecr.eu-west-1.amazonaws.com
    region: eu-west-1
targets:
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
	assert.Equal(t, "eu-west-1", cfg.CurrentRegistry().(*registry.ECR).Region)
	assert.Equal(t, "1234.dkr.ecr.eu-west-1.amazonaws.com", cfg.CurrentRegistry().(*registry.ECR).Url)
	assert.Equal(t, 2, len(cfg.Targets))
	assert.Equal(t, Target{Context: "docker-desktop"}, cfg.Targets["local"])
	devEnv := Target{Context: "docker-desktop", Namespace: "dev"}
	assert.Equal(t, devEnv, cfg.Targets["dev"])

	currentEnv, err := cfg.CurrentTarget("dev")
	assert.NoError(t, err)
	assert.Equal(t, &devEnv, currentEnv)
	_, err = cfg.CurrentTarget("missing")
	assert.EqualError(t, err, "no target matching missing found")
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s/.buildtools.yaml'\x1b[39m\x1b[0m\n", name), out.String())
}

func TestLoad_Old_BrokenYAML(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `ci: []
environments:
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig")
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, ci.No{}.Name(), cfg.CurrentCI().Name())
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s/.buildtools.yaml'\x1b[39m\x1b[0m\n\x1b[0mfile: \x1b[32m'%s/.buildtools.yaml'\x1b[39m \x1b[31mcontains deprecated 'environments' tag, please change to 'targets'\x1b[39m\x1b[0m\n", name, name), out.String())
}

func TestLoad_Old_YAML(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
registry:
  ecr:
    url: 1234.dkr.ecr.eu-west-1.amazonaws.com
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
	assert.Equal(t, "eu-west-1", cfg.CurrentRegistry().(*registry.ECR).Region)
	assert.Equal(t, "1234.dkr.ecr.eu-west-1.amazonaws.com", cfg.CurrentRegistry().(*registry.ECR).Url)
	assert.Equal(t, 2, len(cfg.Targets))
	assert.Equal(t, Target{Context: "docker-desktop"}, cfg.Targets["local"])
	devEnv := Target{Context: "docker-desktop", Namespace: "dev"}
	assert.Equal(t, devEnv, cfg.Targets["dev"])

	currentEnv, err := cfg.CurrentTarget("dev")
	assert.NoError(t, err)
	assert.Equal(t, &devEnv, currentEnv)
	_, err = cfg.CurrentTarget("missing")
	assert.EqualError(t, err, "no target matching missing found")
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s/.buildtools.yaml'\x1b[39m\x1b[0m\n\x1b[0mfile: \x1b[32m'%s/.buildtools.yaml'\x1b[39m \x1b[31mcontains deprecated 'environments' tag, please change to 'targets'\x1b[39m\x1b[0m\n", name, name), out.String())
}

func TestLoad_BrokenYAML_From_Env(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `ci: []
`
	defer pkg.SetEnv("BUILDTOOLS_CONTENT", base64.StdEncoding.EncodeToString([]byte(yaml)))()

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig")
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, ci.No{}.Name(), cfg.CurrentCI().Name())
	assert.NotNil(t, cfg.Registry)
	assert.Equal(t, "Parsing config from env: BUILDTOOLS_CONTENT\n", out.String())
}

func TestLoad_YAML_From_Env_Invalid_Base64(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  local:
    context: docker-desktop
`
	defer pkg.SetEnv("BUILDTOOLS_CONTENT", yaml)()

	out := &bytes.Buffer{}
	_, err := Load(name, out)
	assert.Error(t, err)
	assert.EqualError(t, err, "Failed to decode content: illegal base64 data at input byte 8")
}

func TestLoad_YAML_From_Env(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  local:
    context: docker-desktop
  dev:
    context: docker-desktop
    namespace: dev
`
	defer pkg.SetEnv("BUILDTOOLS_CONTENT", base64.StdEncoding.EncodeToString([]byte(yaml)))()

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 2, len(cfg.Targets))
	assert.Equal(t, Target{Context: "docker-desktop"}, cfg.Targets["local"])
	devEnv := Target{Context: "docker-desktop", Namespace: "dev"}
	assert.Equal(t, devEnv, cfg.Targets["dev"])

	currentEnv, err := cfg.CurrentTarget("dev")
	assert.NoError(t, err)
	assert.Equal(t, &devEnv, currentEnv)
	_, err = cfg.CurrentTarget("missing")
	assert.EqualError(t, err, "no target matching missing found")
	assert.Equal(t, "Parsing config from env: BUILDTOOLS_CONTENT\n", out.String())
}

func TestLoad_Old_YAML_From_Env(t *testing.T) {
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
	defer pkg.SetEnv("BUILDTOOLS_CONTENT", base64.StdEncoding.EncodeToString([]byte(yaml)))()

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 2, len(cfg.Targets))
	assert.Equal(t, Target{Context: "docker-desktop"}, cfg.Targets["local"])
	devEnv := Target{Context: "docker-desktop", Namespace: "dev"}
	assert.Equal(t, devEnv, cfg.Targets["dev"])

	currentEnv, err := cfg.CurrentTarget("dev")
	assert.NoError(t, err)
	assert.Equal(t, &devEnv, currentEnv)
	_, err = cfg.CurrentTarget("missing")
	assert.EqualError(t, err, "no target matching missing found")
	assert.Equal(t, "Parsing config from env: BUILDTOOLS_CONTENT\nBUILDTOOLS_CONTENT contains deprecated 'environments' tag, please change to 'targets'\n", out.String())
}

func TestLoad_YAML_DirStructure(t *testing.T) {
	os.Clearenv()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
registry:
  dockerhub:
    namespace: test
targets:
  test:
    context: abc
  local:
    context: def
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)
	subdir := "sub"
	_ = os.Mkdir(filepath.Join(name, subdir), 0777)
	yaml2 := `
targets:
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
	currentRegistry := cfg.CurrentRegistry()
	assert.NotNil(t, currentRegistry)
	assert.Equal(t, 2, len(cfg.Targets))
	assert.Equal(t, "ghi", cfg.Targets["test"].Context)
	assert.Equal(t, "def", cfg.Targets["local"].Context)
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s/sub/.buildtools.yaml'\x1b[39m\x1b[0m\n\x1b[0mMerging with config from file: \x1b[32m'%s/.buildtools.yaml'\x1b[39m\x1b[0m\n", name, name), out.String())
}

func TestLoad_YAML_Multiple_Registry(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
registry:
  ecr:
    url: 1234.dkr.ecr.eu-west-1.amazonaws.com
    region: eu-west-1
  dockerhub:
    namespace: dockerhub
targets:
  local:
    context: docker-desktop
  dev:
    context: docker-desktop
    namespace: dev
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	out := &bytes.Buffer{}
	_, err := Load(name, out)
	assert.EqualError(t, err, "registry already defined, please check configuration")
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s/.buildtools.yaml'\x1b[39m\x1b[0m\n", name), out.String())
}
