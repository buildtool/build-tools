package service_setup

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	scaffold2 "gitlab.com/sparetimecoders/build-tools/pkg/config/scaffold"
	"gitlab.com/sparetimecoders/build-tools/pkg/config/scaffold/ci"
	"gitlab.com/sparetimecoders/build-tools/pkg/config/scaffold/vcs"
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
	_ = os.Chdir(name)

	return name
}

func teardown(tempDir string) {
	_ = os.RemoveAll(tempDir)
}

func TestSetup_NoArgs(t *testing.T) {
	out := bytes.Buffer{}

	exitCode := Setup(name, &out)

	assert.Equal(t, -1, exitCode)
	assert.Equal(t, "\x1b[0mUsage: service-setup [options] <name>\n\nFor example \x1b[34m`service-setup --stack go gosvc`\x1b[39m would create a new repository and scaffold it as a Go-project\n\nOptions:\n\x1b[0m", out.String())
}

func TestSetup_NonExistingStack(t *testing.T) {
	out := bytes.Buffer{}

	exitCode := Setup(name, &out, "-s", "missing", "project")

	assert.Equal(t, -2, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[31mProvided stack does not exist yet. Available stacks are: \x1b[39m\x1b[97m\x1b[1m(go, none, scala)\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m", out.String())
}

func TestSetup_BrokenConfig(t *testing.T) {
	os.Clearenv()
	yaml := `ci: [] `
	file := filepath.Join(name, ".buildtools.yaml")
	_ = ioutil.WriteFile(file, []byte(yaml), 0777)
	defer func() { _ = os.Remove(file) }()

	out := bytes.Buffer{}

	exitCode := Setup(name, &out, "project")

	assert.Equal(t, -3, exitCode)
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s'\x1b[39m\x1b[0m\n\x1b[0m\x1b[31myaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig\x1b[39m\x1b[0m\n", file), out.String())
}

func TestSetup_NoVCS(t *testing.T) {
	out := bytes.Buffer{}

	exitCode := Setup(name, &out, "project")

	assert.Equal(t, -4, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[31mno VCS configured\x1b[39m\x1b[0m\n", out.String())
}

func TestScaffold_Missing_Token(t *testing.T) {
	yaml := `
scaffold:
  ci:
    selected: buildkite
    buildkite:
      organisation: example
      token: abc
  vcs:
    selected: github
    github:
      organisation: example
      token: abc
`
	file := filepath.Join(name, ".buildtools.yaml")
	_ = ioutil.WriteFile(file, []byte(yaml), 0777)
	defer func() { _ = os.Remove(file) }()

	out := bytes.Buffer{}

	exitCode := Setup(name, &out, "project")

	assert.Equal(t, -6, exitCode)
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s'\x1b[39m\x1b[0m\n\x1b[0m\x1b[31mGET https://api.buildkite.com/v2/user: 401 Authentication required. Please supply a valid API Access Token: https://buildkite.com/docs/apis/rest-api#authentication\x1b[39m\x1b[0m\n", file), out.String())
}

func TestScaffold_Configure_Error(t *testing.T) {
	cfg := scaffold2.InitEmptyConfig()
	cfg.VCS.Selected = "mock"
	cfg.CI.Selected = "mock"
	cfg.CurrentCI = &mockCi{configErr: errors.New("config error")}
	cfg.CurrentVCS = &mockVcs{}
	out := &bytes.Buffer{}
	exitCode := scaffold(cfg, name, "project", &stack.None{}, out)
	assert.Equal(t, -5, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[31mconfig error\x1b[39m\x1b[0m\n", out.String())
}

func TestScaffold_Ok(t *testing.T) {
	cfg := scaffold2.InitEmptyConfig()
	cfg.VCS.Selected = "mock"
	cfg.CI.Selected = "mock"
	cfg.CurrentCI = &mockCi{}
	cfg.CurrentVCS = &mockVcs{}
	out := &bytes.Buffer{}
	exitCode := scaffold(cfg, name, "project", &stack.None{}, out)
	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'none'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mock'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'git@git'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m", out.String())
}

type mockCi struct {
	configErr error
}

func (m mockCi) Name() string {
	panic("implement me")
}

func (m mockCi) ValidateConfig() error {
	panic("implement me")
}

func (m mockCi) Validate(name string) error {
	return nil
}

func (m mockCi) Scaffold(dir string, data templating.TemplateData) (*string, error) {
	return nil, nil
}

func (m mockCi) Badges(name string) ([]templating.Badge, error) {
	return nil, nil
}

func (m mockCi) Configure() error {
	return m.configErr
}

var _ ci.CI = &mockCi{}

type mockVcs struct {
}

func (m mockVcs) Name() string {
	return "mock"
}

func (m mockVcs) ValidateConfig() error {
	panic("implement me")
}

func (m mockVcs) Configure() {
}

func (m mockVcs) Validate(name string) error {
	return nil
}

func (m mockVcs) Scaffold(name string) (*vcs.RepositoryInfo, error) {
	return &vcs.RepositoryInfo{
		SSHURL:  "git@git",
		HTTPURL: "https://git",
	}, nil
}

func (m mockVcs) Webhook(name, url string) error {
	panic("implement me")
}

func (m mockVcs) Clone(dir, name, url string, out io.Writer) error {
	return nil
}

var _ vcs.VCS = &mockVcs{}
