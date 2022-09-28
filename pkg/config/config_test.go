// MIT License
//
// Copyright (c) 2021 buildtool
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg"
	"github.com/buildtool/build-tools/pkg/ci"
	"github.com/buildtool/build-tools/pkg/registry"
)

var name string

func TestMain(m *testing.M) {
	tempDir := setup()
	code := m.Run()
	teardown(tempDir)
	os.Exit(code)
}

func setup() string {
	name, _ = os.MkdirTemp(os.TempDir(), "build-tools")

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

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	_, err := Load("test")
	assert.EqualError(t, err, "abs-error")
	logMock.Check(t, []string{})
	abs = filepath.Abs
}

func TestLoad_Empty(t *testing.T) {
	os.Clearenv()
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	cfg, err := Load(name)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, ci.No{}.Name(), cfg.CurrentCI().Name())
	assert.NotNil(t, cfg.Registry)
	logMock.Check(t, []string{})
}

func TestLoad_BrokenYAML(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `ci: []
`
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	cfg, err := Load(name)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig")
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, ci.No{}.Name(), cfg.CurrentCI().Name())
	assert.NotNil(t, cfg.Registry)
	logMock.Check(t, []string{fmt.Sprintf("debug: Parsing config from file: <green>'%s/.buildtools.yaml'</green>\n", name)})
}

func TestLoad_UnreadableFile(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	filename := filepath.Join(name, ".buildtools.yaml")
	_ = os.Mkdir(filename, 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	cfg, err := Load(name)
	assert.EqualError(t, err, fmt.Sprintf("read %s: is a directory", filename))
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, ci.No{}.Name(), cfg.CurrentCI().Name())
	assert.NotNil(t, cfg.Registry)
	logMock.Check(t, []string{fmt.Sprintf("debug: Parsing config from file: <green>'%s/.buildtools.yaml'</green>\n", name)})
}

func TestLoad_YAML(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
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
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	cfg, err := Load(name)
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
	logMock.Check(t, []string{fmt.Sprintf("debug: Parsing config from file: <green>'%s/.buildtools.yaml'</green>\n", name)})
}

func TestLoad_BrokenYAML_From_Env(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `ci: []
`
	defer pkg.SetEnv(envBuildtoolsContent, base64.StdEncoding.EncodeToString([]byte(yaml)))()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	cfg, err := Load(name)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig")
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.CI)
	assert.Equal(t, ci.No{}.Name(), cfg.CurrentCI().Name())
	assert.NotNil(t, cfg.Registry)
	logMock.Check(t, []string{"debug: Parsing config from env: BUILDTOOLS_CONTENT\n"})
}

func TestLoad_YAML_From_Env_Plain(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  local:
    context: docker-desktop
`
	defer pkg.SetEnv(envBuildtoolsContent, yaml)()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	cfg, err := Load(name)
	assert.NoError(t, err)
	logMock.Check(t, []string{"debug: Parsing config from env: BUILDTOOLS_CONTENT\n",
		"debug: Failed to decode BASE64, falling back to plaintext\n"})
	assert.Equal(t, len(cfg.Targets), 1)
	assert.Equal(t, cfg.Targets["local"].Context, "docker-desktop")
}

func TestLoad_Broken_YAML_From_Env_Plain(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
target:
  local:
    context: docker-desktop
`
	defer pkg.SetEnv(envBuildtoolsContent, yaml)()

	_, err := Load(name)
	assert.Error(t, err)
	assert.EqualError(t, err, "yaml: unmarshal errors:\n  line 2: field target not found in type config.Config")
}

func TestLoad_YAML_From_Env(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  local:
    context: docker-desktop
  dev:
    context: docker-desktop
    namespace: dev
`
	defer pkg.SetEnv(envBuildtoolsContent, base64.StdEncoding.EncodeToString([]byte(yaml)))()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	cfg, err := Load(name)
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
	logMock.Check(t, []string{"debug: Parsing config from env: BUILDTOOLS_CONTENT\n"})
}

func TestLoad_YAML_DirStructure(t *testing.T) {
	os.Clearenv()
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
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
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)
	subdir := "sub"
	_ = os.Mkdir(filepath.Join(name, subdir), 0777)
	yaml2 := `
targets:
  test:
    context: ghi
`
	_ = os.WriteFile(filepath.Join(name, subdir, ".buildtools.yaml"), []byte(yaml2), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	cfg, err := Load(filepath.Join(name, subdir))
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
	logMock.Check(t, []string{fmt.Sprintf("debug: Parsing config from file: <green>'%s/sub/.buildtools.yaml'</green>\n", name),
		fmt.Sprintf("debug: Merging with config from file: <green>'%s/.buildtools.yaml'</green>\n", name)})
}

func TestLoad_YAML_Multiple_Registry(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
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
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	_, err := Load(name)
	assert.EqualError(t, err, "registry already defined, please check configuration")
	logMock.Check(t, []string{fmt.Sprintf("debug: Parsing config from file: <green>'%s/.buildtools.yaml'</green>\n", name)})
}
