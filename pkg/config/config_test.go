// MIT License
//
// Copyright (c) 2018 buildtool
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
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0o777)

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
	_ = os.Mkdir(filename, 0o777)

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
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0o777)

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
	logMock.Check(t, []string{
		"debug: Parsing config from env: BUILDTOOLS_CONTENT\n",
		"debug: Failed to decode BASE64, falling back to plaintext\n",
	})
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
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0o777)
	subdir := "sub"
	_ = os.Mkdir(filepath.Join(name, subdir), 0o777)
	yaml2 := `
targets:
  test:
    context: ghi
`
	_ = os.WriteFile(filepath.Join(name, subdir, ".buildtools.yaml"), []byte(yaml2), 0o777)

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
	logMock.Check(t, []string{
		fmt.Sprintf("debug: Parsing config from file: <green>'%s/sub/.buildtools.yaml'</green>\n", name),
		fmt.Sprintf("debug: Merging with config from file: <green>'%s/.buildtools.yaml'</green>\n", name),
	})
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
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0o777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	_, err := Load(name)
	assert.EqualError(t, err, "registry already defined, please check configuration")
	logMock.Check(t, []string{fmt.Sprintf("debug: Parsing config from file: <green>'%s/.buildtools.yaml'</green>\n", name)})
}

func TestECRCache_Configured(t *testing.T) {
	tests := []struct {
		name  string
		cache *ECRCache
		want  bool
	}{
		{
			name:  "nil cache",
			cache: nil,
			want:  false,
		},
		{
			name:  "empty cache",
			cache: &ECRCache{},
			want:  false,
		},
		{
			name:  "url only",
			cache: &ECRCache{Url: "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache"},
			want:  true,
		},
		{
			name:  "url and tag",
			cache: &ECRCache{Url: "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache", Tag: "custom"},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cache.Configured())
		})
	}
}

func TestECRCache_CacheRef(t *testing.T) {
	tests := []struct {
		name  string
		cache *ECRCache
		want  string
	}{
		{
			name:  "default tag",
			cache: &ECRCache{Url: "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache"},
			want:  "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache:buildcache",
		},
		{
			name:  "custom tag",
			cache: &ECRCache{Url: "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache", Tag: "v1"},
			want:  "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache:v1",
		},
		{
			name:  "empty tag uses default",
			cache: &ECRCache{Url: "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache", Tag: ""},
			want:  "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache:buildcache",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.cache.CacheRef())
		})
	}
}

func TestECRCache_AsRegistry(t *testing.T) {
	tests := []struct {
		name       string
		cache      *ECRCache
		wantNil    bool
		wantUrl    string
		wantRegion string
	}{
		{
			name:    "nil cache",
			cache:   nil,
			wantNil: true,
		},
		{
			name:    "empty cache",
			cache:   &ECRCache{},
			wantNil: true,
		},
		{
			name:       "configured cache us-east-1",
			cache:      &ECRCache{Url: "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache"},
			wantNil:    false,
			wantUrl:    "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache",
			wantRegion: "us-east-1",
		},
		{
			name:       "configured cache eu-west-1",
			cache:      &ECRCache{Url: "987654321098.dkr.ecr.eu-west-1.amazonaws.com/my-cache"},
			wantNil:    false,
			wantUrl:    "987654321098.dkr.ecr.eu-west-1.amazonaws.com/my-cache",
			wantRegion: "eu-west-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := tt.cache.AsRegistry()
			if tt.wantNil {
				assert.Nil(t, reg)
			} else {
				assert.NotNil(t, reg)
				assert.Equal(t, tt.wantUrl, reg.RegistryUrl())
				assert.Equal(t, tt.wantRegion, reg.Region)
			}
		})
	}
}

func TestECRCache_extractRegion(t *testing.T) {
	tests := []struct {
		name  string
		cache *ECRCache
		want  string
	}{
		{
			name:  "nil cache",
			cache: nil,
			want:  "",
		},
		{
			name:  "empty url",
			cache: &ECRCache{Url: ""},
			want:  "",
		},
		{
			name:  "us-east-1",
			cache: &ECRCache{Url: "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache"},
			want:  "us-east-1",
		},
		{
			name:  "eu-west-1",
			cache: &ECRCache{Url: "987654321098.dkr.ecr.eu-west-1.amazonaws.com/cache"},
			want:  "eu-west-1",
		},
		{
			name:  "ap-southeast-2",
			cache: &ECRCache{Url: "111222333444.dkr.ecr.ap-southeast-2.amazonaws.com/my-cache-repo"},
			want:  "ap-southeast-2",
		},
		{
			name:  "invalid url format",
			cache: &ECRCache{Url: "docker.io/library/alpine"},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cache.extractRegion()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestECRCache_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cache   *ECRCache
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil cache",
			cache:   nil,
			wantErr: false,
		},
		{
			name:    "empty url",
			cache:   &ECRCache{Url: ""},
			wantErr: false,
		},
		{
			name:    "valid ECR URL us-east-1",
			cache:   &ECRCache{Url: "123456789012.dkr.ecr.us-east-1.amazonaws.com/cache"},
			wantErr: false,
		},
		{
			name:    "valid ECR URL eu-west-1",
			cache:   &ECRCache{Url: "987654321098.dkr.ecr.eu-west-1.amazonaws.com/my-cache"},
			wantErr: false,
		},
		{
			name:    "valid ECR URL without repo path",
			cache:   &ECRCache{Url: "123456789012.dkr.ecr.us-west-2.amazonaws.com"},
			wantErr: false,
		},
		{
			name:    "valid ECR URL with nested repo path",
			cache:   &ECRCache{Url: "123456789012.dkr.ecr.eu-central-1.amazonaws.com/org/project/cache"},
			wantErr: false,
		},
		{
			name:    "invalid - Docker Hub URL",
			cache:   &ECRCache{Url: "docker.io/library/alpine"},
			wantErr: true,
			errMsg:  "invalid ECR cache URL",
		},
		{
			name:    "invalid - GitHub Container Registry",
			cache:   &ECRCache{Url: "ghcr.io/owner/repo"},
			wantErr: true,
			errMsg:  "invalid ECR cache URL",
		},
		{
			name:    "invalid - GitLab Registry",
			cache:   &ECRCache{Url: "registry.gitlab.com/group/project"},
			wantErr: true,
			errMsg:  "invalid ECR cache URL",
		},
		{
			name:    "invalid - account ID too short",
			cache:   &ECRCache{Url: "12345.dkr.ecr.us-east-1.amazonaws.com/cache"},
			wantErr: true,
			errMsg:  "invalid ECR cache URL",
		},
		{
			name:    "invalid - account ID too long",
			cache:   &ECRCache{Url: "1234567890123.dkr.ecr.us-east-1.amazonaws.com/cache"},
			wantErr: true,
			errMsg:  "invalid ECR cache URL",
		},
		{
			name:    "invalid - account ID with letters",
			cache:   &ECRCache{Url: "12345678901a.dkr.ecr.us-east-1.amazonaws.com/cache"},
			wantErr: true,
			errMsg:  "invalid ECR cache URL",
		},
		{
			name:    "invalid - missing dkr",
			cache:   &ECRCache{Url: "123456789012.ecr.us-east-1.amazonaws.com/cache"},
			wantErr: true,
			errMsg:  "invalid ECR cache URL",
		},
		{
			name:    "invalid - missing ecr",
			cache:   &ECRCache{Url: "123456789012.dkr.us-east-1.amazonaws.com/cache"},
			wantErr: true,
			errMsg:  "invalid ECR cache URL",
		},
		{
			name:    "invalid - wrong domain",
			cache:   &ECRCache{Url: "123456789012.dkr.ecr.us-east-1.example.com/cache"},
			wantErr: true,
			errMsg:  "invalid ECR cache URL",
		},
		{
			name:    "invalid - S3 URL",
			cache:   &ECRCache{Url: "s3://my-bucket/cache"},
			wantErr: true,
			errMsg:  "invalid ECR cache URL",
		},
		{
			name:    "invalid - random string",
			cache:   &ECRCache{Url: "not-a-valid-url"},
			wantErr: true,
			errMsg:  "invalid ECR cache URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cache.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoad_InvalidECRCacheURL(t *testing.T) {
	yaml := `
cache:
  ecr:
    url: docker.io/library/alpine
    tag: cache
`
	name := filepath.Join(t.TempDir(), ".buildtools.yaml")
	_ = os.WriteFile(name, []byte(yaml), 0o644)

	_, err := Load(filepath.Dir(name))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ECR cache URL")
}

func TestLoad_UnknownCacheType(t *testing.T) {
	yaml := `
cache:
  gitlab:
    url: registry.gitlab.com/group/project
`
	name := filepath.Join(t.TempDir(), ".buildtools.yaml")
	_ = os.WriteFile(name, []byte(yaml), 0o644)

	_, err := Load(filepath.Dir(name))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gitlab")
}

func TestLoad_YAML_Gitea(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
registry:
  gitea:
    registry: gitea.example.com
    username: user
    token: token
    repository: org/repo
targets:
  local:
    context: docker-desktop
`
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0o777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	cfg, err := Load(name)
	assert.Nil(t, err)
	assert.True(t, cfg.Registry.Gitea.Configured())
	assert.Equal(t, "Gitea", cfg.CurrentRegistry().Name())
	assert.Equal(t, "gitea.example.com", cfg.Registry.Gitea.Registry)
	assert.Equal(t, "user", cfg.Registry.Gitea.Username)
	assert.Equal(t, "token", cfg.Registry.Gitea.Token)
	assert.Equal(t, "org/repo", cfg.Registry.Gitea.Repository)
	assert.Equal(t, "gitea.example.com/org", cfg.CurrentRegistry().RegistryUrl())
	logMock.Check(t, []string{fmt.Sprintf("debug: Parsing config from file: <green>'%s/.buildtools.yaml'</green>\n", name)})
}

func TestLoad_YAML_Gitea_Env(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	t.Setenv("GITEA_REGISTRY", "gitea.example.com")
	t.Setenv("GITEA_USERNAME", "user")
	t.Setenv("GITEA_TOKEN", "token")
	t.Setenv("GITEA_REPOSITORY", "org/repo")

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	cfg, err := Load(name)
	assert.Nil(t, err)
	assert.True(t, cfg.Registry.Gitea.Configured())
	assert.Equal(t, "Gitea", cfg.CurrentRegistry().Name())
	assert.Equal(t, "gitea.example.com", cfg.Registry.Gitea.Registry)
	assert.Equal(t, "user", cfg.Registry.Gitea.Username)
	assert.Equal(t, "token", cfg.Registry.Gitea.Token)
	assert.Equal(t, "org/repo", cfg.Registry.Gitea.Repository)
}
