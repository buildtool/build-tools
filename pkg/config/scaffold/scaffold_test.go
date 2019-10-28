package scaffold

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sparetimecoders/build-tools/pkg"
	"github.com/sparetimecoders/build-tools/pkg/config/scaffold/ci"
	"github.com/sparetimecoders/build-tools/pkg/config/scaffold/vcs"
	"github.com/sparetimecoders/build-tools/pkg/stack"
	"github.com/sparetimecoders/build-tools/pkg/templating"
	"github.com/stretchr/testify/assert"
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

func TestValidateConfig_No_VCS_Configured(t *testing.T) {
	cfg := InitEmptyConfig()
	err := cfg.ValidateConfig()
	assert.EqualError(t, err, "no VCS configured")
}

func TestValidateConfig_No_CI_Configured(t *testing.T) {
	cfg := InitEmptyConfig()
	cfg.CurrentVCS = mockVcs{}
	err := cfg.ValidateConfig()
	assert.EqualError(t, err, "no CI configured")
}

func TestConfigure(t *testing.T) {
	cfg := InitEmptyConfig()
	cfg.CurrentCI = &mockCi{}
	cfg.CurrentVCS = &mockVcs{}

	err := cfg.Configure()

	assert.NoError(t, err)
}

func TestValidate_VCS_Error(t *testing.T) {
	cfg := InitEmptyConfig()
	cfg.CurrentCI = &mockCi{}
	cfg.CurrentVCS = &mockVcs{validateErr: errors.New("validate error")}

	err := cfg.Validate("project")

	assert.EqualError(t, err, "validate error")
}

func TestValidate_CI_Error(t *testing.T) {
	cfg := InitEmptyConfig()
	cfg.CurrentCI = &mockCi{validateErr: errors.New("validate error")}
	cfg.CurrentVCS = &mockVcs{}

	err := cfg.Validate("project")

	assert.EqualError(t, err, "validate error")
}

func TestScaffold_VcsScaffold_Error(t *testing.T) {
	cfg := InitEmptyConfig()
	cfg.CurrentVCS = &mockVcs{scaffoldErr: errors.New("error")}
	cfg.RegistryUrl = "dockerhub"

	out := &bytes.Buffer{}

	exitCode := cfg.Scaffold(name, "project", &stack.None{}, out)

	assert.Equal(t, -7, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'none'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31merror\x1b[39m\x1b[0m\n", out.String())
}

func TestScaffold_VcsClone_Error(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := InitEmptyConfig()
	cfg.CurrentVCS = &mockVcs{cloneErr: errors.New("error")}
	cfg.RegistryUrl = "dockerhub"

	out := &bytes.Buffer{}

	exitCode := cfg.Scaffold(name, "project", &stack.None{}, out)

	assert.Equal(t, -8, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'none'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31merror\x1b[39m\x1b[0m\n", out.String())
}

func TestScaffold_Badges_Error(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := InitEmptyConfig()
	cfg.CurrentVCS = &mockVcs{}
	cfg.CurrentCI = &mockCi{badgesErr: errors.New("error")}
	cfg.RegistryUrl = "dockerhub"

	out := &bytes.Buffer{}

	exitCode := cfg.Scaffold(name, "project", &stack.None{}, out)

	assert.Equal(t, -9, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'none'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31merror\x1b[39m\x1b[0m\n", out.String())
}

func TestScaffold_Not_Parsable_Repository_Url(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := InitEmptyConfig()
	cfg.CurrentVCS = &mockVcs{httpUrl: "http://192.168.0.%31/"}
	cfg.CurrentCI = &mockCi{}
	cfg.RegistryUrl = "dockerhub"

	out := &bytes.Buffer{}

	exitCode := cfg.Scaffold(name, "project", &stack.None{}, out)

	assert.Equal(t, -10, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'none'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31mparse http://192.168.0.%31/: invalid URL escape \"%31\"\x1b[39m\x1b[0m\n", out.String())
}

func TestScaffold_CiScaffold_Error(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := InitEmptyConfig()
	cfg.CurrentVCS = &mockVcs{}
	cfg.CurrentCI = &mockCi{scaffoldErr: errors.New("error")}
	cfg.RegistryUrl = "dockerhub"

	out := &bytes.Buffer{}

	exitCode := cfg.Scaffold(name, "project", &stack.None{}, out)

	assert.Equal(t, -11, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'none'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31merror\x1b[39m\x1b[0m\n", out.String())
}

func TestScaffold_Webhook_Error(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := InitEmptyConfig()
	cfg.CurrentVCS = &mockVcs{webhookErr: errors.New("error")}
	cfg.CurrentCI = &mockCi{webhookUrl: pkg.String("https://example.org")}
	cfg.RegistryUrl = "dockerhub"

	out := &bytes.Buffer{}

	exitCode := cfg.Scaffold(name, "project", &stack.None{}, out)

	assert.Equal(t, -12, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'none'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31merror\x1b[39m\x1b[0m\n", out.String())
}

func TestScaffold_Error_Writing_Gitignore(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := InitEmptyConfig()
	cfg.CurrentVCS = &mockVcs{}
	cfg.CurrentCI = &mockCi{}
	cfg.RegistryUrl = "dockerhub"

	filename := filepath.Join(name, "project", ".gitignore")
	_ = os.MkdirAll(filename, 0777)
	out := &bytes.Buffer{}

	exitCode := cfg.Scaffold(name, "project", &errorStack{}, out)
	assert.Equal(t, -13, exitCode)

	assert.Equal(t, fmt.Sprintf("\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'error-stack'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31mopen %s: is a directory\x1b[39m\x1b[0m\n", filename), out.String())
}
func TestScaffold_Error_Writing_Editorconfig(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := InitEmptyConfig()
	cfg.CurrentVCS = &mockVcs{}
	cfg.CurrentCI = &mockCi{}
	cfg.RegistryUrl = "dockerhub"

	filename := filepath.Join(name, "project", ".editorconfig")
	_ = os.MkdirAll(filename, 0777)
	out := &bytes.Buffer{}

	exitCode := cfg.Scaffold(name, "project", &errorStack{}, out)
	assert.Equal(t, -13, exitCode)

	assert.Equal(t, fmt.Sprintf("\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'error-stack'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31mopen %s: is a directory\x1b[39m\x1b[0m\n", filename), out.String())
}

func TestScaffold_Error_Writing_Dockerignore(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := InitEmptyConfig()
	cfg.CurrentVCS = &mockVcs{}
	cfg.CurrentCI = &mockCi{}
	cfg.RegistryUrl = "dockerhub"

	filename := filepath.Join(name, "project", ".dockerignore")
	_ = os.MkdirAll(filename, 0777)
	out := &bytes.Buffer{}

	exitCode := cfg.Scaffold(name, "project", &errorStack{}, out)
	assert.Equal(t, -13, exitCode)

	assert.Equal(t, fmt.Sprintf("\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'error-stack'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31mopen %s: is a directory\x1b[39m\x1b[0m\n", filename), out.String())
}

func TestScaffold_Error_Writing_Readme(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := InitEmptyConfig()
	cfg.CurrentVCS = &mockVcs{}
	cfg.CurrentCI = &mockCi{}
	cfg.RegistryUrl = "dockerhub"

	filename := filepath.Join(name, "project", "README.md")
	_ = os.MkdirAll(filename, 0777)
	out := &bytes.Buffer{}

	exitCode := cfg.Scaffold(name, "project", &errorStack{}, out)
	assert.Equal(t, -14, exitCode)

	assert.Equal(t, fmt.Sprintf("\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'error-stack'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31mopen %s: is a directory\x1b[39m\x1b[0m\n", filename), out.String())
}

func TestScaffold_Error_Writing_Deployment(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := InitEmptyConfig()
	cfg.CurrentVCS = &mockVcs{}
	cfg.CurrentCI = &mockCi{}
	cfg.RegistryUrl = "dockerhub"

	filename := filepath.Join(name, "project", "k8s", "deploy.yaml")
	_ = os.MkdirAll(filename, 0777)
	out := &bytes.Buffer{}

	exitCode := cfg.Scaffold(name, "project", &errorStack{}, out)
	assert.Equal(t, -15, exitCode)

	assert.Equal(t, fmt.Sprintf("\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'error-stack'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31mopen %s: is a directory\x1b[39m\x1b[0m\n", filename), out.String())
}

func TestScaffold_StackError(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := InitEmptyConfig()
	cfg.CurrentVCS = &mockVcs{}
	cfg.CurrentCI = &mockCi{}
	cfg.RegistryUrl = "dockerhub"

	out := &bytes.Buffer{}

	exitCode := cfg.Scaffold(name, "project", &errorStack{}, out)
	assert.Equal(t, -16, exitCode)

	assert.Equal(t, "\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'error-stack'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[31merror\x1b[39m\x1b[0m\n", out.String())
}

func TestScaffold_Ok(t *testing.T) {
	defer func() { _ = os.RemoveAll(name) }()
	os.Clearenv()
	cfg := InitEmptyConfig()
	cfg.CurrentVCS = &mockVcs{}
	cfg.CurrentCI = &mockCi{}
	cfg.RegistryUrl = "dockerhub"

	out := &bytes.Buffer{}

	exitCode := cfg.Scaffold(name, "project", &stack.None{}, out)
	assert.Equal(t, 0, exitCode)

	assert.Equal(t, "\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'none'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'mockVcs'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m'file:///tmp'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m", out.String())
}

type errorStack struct{}

func (e errorStack) Scaffold(dir string, data templating.TemplateData) error {
	return errors.New("error")
}

func (e errorStack) Name() string {
	return "error-stack"
}

var _ stack.Stack = &errorStack{}

type mockCi struct {
	webhookUrl  *string
	validateErr error
	badgesErr   error
	scaffoldErr error
}

func (m mockCi) Name() string {
	panic("implement me")
}

func (m mockCi) ValidateConfig() error {
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

func (m mockCi) Validate(name string) error {
	return m.validateErr
}

func (m mockCi) Scaffold(dir string, data templating.TemplateData) (*string, error) {
	if m.scaffoldErr != nil {
		return nil, m.scaffoldErr
	}
	return m.webhookUrl, nil
}

func (m mockCi) Badges(name string) ([]templating.Badge, error) {
	return nil, m.badgesErr
}

func (m mockCi) Configure() error {
	return nil
}

func (m mockCi) Configured() bool {
	return true
}

var _ ci.CI = &mockCi{}

type mockVcs struct {
	skipMkdir   bool
	validateErr error
	scaffoldErr error
	cloneErr    error
	webhookErr  error
	httpUrl     string
}

func (m mockVcs) Name() string {
	return "mockVcs"
}

func (m mockVcs) ValidateConfig() error {
	panic("implement me")
}

func (m mockVcs) Configure() {
}

func (m mockVcs) Validate(name string) error {
	return m.validateErr
}

func (m mockVcs) Scaffold(name string) (*vcs.RepositoryInfo, error) {
	if m.scaffoldErr != nil {
		return nil, m.scaffoldErr
	}
	return &vcs.RepositoryInfo{
		SSHURL:  "file:///tmp",
		HTTPURL: m.httpUrl,
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

var _ vcs.VCS = &mockVcs{}
