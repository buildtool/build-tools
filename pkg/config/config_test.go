package config

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/stack"
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"io"
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
	cfg, err := Load(name, out)
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
	cfg, err := Load(name, out)
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
	cfg, err := Load(name, out)
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
	cfg, err := Load(name, out)
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
	cfg, err := Load(name, out)
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
	cfg, err := Load(name, out)
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
	cfg, err := Load(name, out)
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
	cfg, err := Load(name, out)
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

func TestScaffold_Registry_Error(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := initEmptyConfig()
	cfg.VCS.Selected = "github"
	cfg.VCS.Github = &GithubVCS{}

	out := &bytes.Buffer{}
	exitCode := 0

	cfg.Scaffold(name, "project", &stack.None{}, out, func(code int) {
		exitCode = code
	})

	assert.Equal(t, -2, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[31mno Docker registry found\x1b[39m\x1b[0m\n", out.String())
}

func TestScaffold_Validate_Error(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := initEmptyConfig()
	cfg.VCS.Selected = "github"
	cfg.VCS.Github = &GithubVCS{}
	cfg.Registry.Selected = "dockerhub"

	out := &bytes.Buffer{}
	exitCode := 0

	cfg.Scaffold(name, "project", &stack.None{}, out, func(code int) {
		exitCode = code
	})

	assert.Equal(t, -3, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[31mtoken is required\x1b[39m\x1b[0m\n", out.String())
}

func TestScaffold_VcsScaffold_Error(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := initEmptyConfig()
	cfg.VCS.VCS = &mockVcs{scaffoldErr: errors.New("error")}
	cfg.Registry.Selected = "dockerhub"

	out := &bytes.Buffer{}
	exitCode := 0

	cfg.Scaffold(name, "project", &stack.None{}, out, func(code int) {
		exitCode = code
	})

	assert.Equal(t, -4, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'none'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31merror\x1b[39m\x1b[0m\n", out.String())
}

func TestScaffold_VcsClone_Error(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := initEmptyConfig()
	cfg.VCS.VCS = &mockVcs{cloneErr: errors.New("error")}
	cfg.Registry.Selected = "dockerhub"

	out := &bytes.Buffer{}
	exitCode := 0

	cfg.Scaffold(name, "project", &stack.None{}, out, func(code int) {
		exitCode = code
	})

	assert.Equal(t, -5, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'none'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31merror\x1b[39m\x1b[0m\n", out.String())
}

func TestScaffold_CreateDirectories_Error(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := initEmptyConfig()
	cfg.VCS.VCS = &mockVcs{skipMkdir: true}
	cfg.availableCI = []CI{&mockCi{scaffoldErr: errors.New("error")}}
	cfg.Registry.Selected = "dockerhub"

	out := &bytes.Buffer{}
	exitCode := 0

	cfg.Scaffold(name, "project", &stack.None{}, out, func(code int) {
		exitCode = code
	})

	assert.Equal(t, -6, exitCode)
	assert.Equal(t, fmt.Sprintf("\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'none'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31mmkdir %s: no such file or directory\x1b[39m\x1b[0m\n", filepath.Join(name, "project", "deployment_files")), out.String())
}

func TestScaffold_CiScaffold_Error(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := initEmptyConfig()
	cfg.VCS.VCS = &mockVcs{}
	cfg.availableCI = []CI{&mockCi{scaffoldErr: errors.New("error")}}
	cfg.Registry.Selected = "dockerhub"

	out := &bytes.Buffer{}
	exitCode := 0

	cfg.Scaffold(name, "project", &stack.None{}, out, func(code int) {
		exitCode = code
	})

	assert.Equal(t, -7, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'none'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31merror\x1b[39m\x1b[0m\n", out.String())
}

func TestScaffold_Webhook_Error(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := initEmptyConfig()
	cfg.VCS.VCS = &mockVcs{webhookErr: errors.New("error")}
	cfg.availableCI = []CI{&mockCi{webhookUrl: wrapString("https://example.org")}}
	cfg.Registry.Selected = "dockerhub"

	out := &bytes.Buffer{}
	exitCode := 0

	cfg.Scaffold(name, "project", &stack.None{}, out, func(code int) {
		exitCode = code
	})

	assert.Equal(t, -8, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'none'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31merror\x1b[39m\x1b[0m\n", out.String())
}

func TestScaffold_StackError(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := initEmptyConfig()
	cfg.VCS.VCS = &mockVcs{}
	cfg.availableCI = []CI{&mockCi{}}
	cfg.Registry.Selected = "dockerhub"

	out := &bytes.Buffer{}
	exitCode := 0

	cfg.Scaffold(name, "project", &errorStack{}, out, func(code int) {
		exitCode = code
	})

	assert.Equal(t, -13, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'error-stack'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31merror\x1b[39m\x1b[0m\n", out.String())
}

type errorStack struct{}

func (e errorStack) Scaffold(dir, name string, data templating.TemplateData) error {
	return errors.New("error")
}

func (e errorStack) Name() string {
	return "error-stack"
}

var _ stack.Stack = &errorStack{}

type mockCi struct {
	ci
	webhookUrl  *string
	scaffoldErr error
}

func (m mockCi) Name() string {
	panic("implement me")
}

func (m mockCi) Branch() string {
	panic("implement me")
}

func (m mockCi) BranchReplaceSlash() string {
	panic("implement me")
}

func (m mockCi) Commit() string {
	panic("implement me")
}

func (m mockCi) Scaffold(dir, name, repository string) (*string, error) {
	if m.scaffoldErr != nil {
		return nil, m.scaffoldErr
	}
	return m.webhookUrl, nil
}

func (m mockCi) Badges() string {
	return ""
}

func (m mockCi) configured() bool {
	return true
}

var _ CI = &mockCi{}

type mockVcs struct {
	skipMkdir   bool
	scaffoldErr error
	cloneErr    error
	webhookErr  error
}

func (m mockVcs) identify(dir string, out io.Writer) bool {
	panic("implement me")
}

func (m mockVcs) Name() string {
	return "mockVcs"
}

func (m mockVcs) Branch() string {
	panic("implement me")
}

func (m mockVcs) Commit() string {
	panic("implement me")
}

func (m mockVcs) Scaffold(name string) (*RepositoryInfo, error) {
	if m.scaffoldErr != nil {
		return nil, m.scaffoldErr
	}
	return &RepositoryInfo{
		SSHURL:  "file:///tmp",
		HTTPURL: "http://github.com/example/repo",
	}, nil
}

func (m mockVcs) Webhook(name, url string) error {
	return m.webhookErr
}

func (m mockVcs) Clone(dir, name, url string, out io.Writer) error {
	if m.cloneErr != nil {
		return m.cloneErr
	}
	if !m.skipMkdir {
		_ = os.MkdirAll(filepath.Join(dir, name), 0777)
	}
	return nil
}

func (m mockVcs) Validate() error {
	return nil
}

var _ VCS = &mockVcs{}
