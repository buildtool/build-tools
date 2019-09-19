package config

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestIdentify_Azure(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("VSTS_PROCESS_LOOKUP_ID", "1")
	_ = os.Setenv("BUILD_SOURCEVERSION", "abc123")
	_ = os.Setenv("BUILD_REPOSITORY_NAME", "reponame")
	_ = os.Setenv("BUILD_SOURCEBRANCHNAME", "feature/first test")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, "abc123", result.Commit())
	assert.Equal(t, "reponame", result.BuildName())
	assert.Equal(t, "feature/first test", result.Branch())
	assert.Equal(t, "feature_first_test", result.BranchReplaceSlash())
	assert.Equal(t, "", out.String())
}

func TestName_Azure(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("VSTS_PROCESS_LOOKUP_ID", "1")
	_ = os.Setenv("BUILD_SOURCEVERSION", "abc123")
	_ = os.Setenv("BUILD_REPOSITORY_NAME", "reponame")
	_ = os.Setenv("BUILD_SOURCEBRANCHNAME", "feature/first test")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.Equal(t, "Azure", result.Name())
}

func TestBuildName_Fallback_Azure(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "azure")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)
	oldPwd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer os.Chdir(oldPwd)

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, filepath.Base(dir), result.BuildName())
	assert.Equal(t, "", out.String())
}

func TestBranch_VCS_Fallback_Azure(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "azure")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, "master", result.Branch())
	assert.Equal(t, "", out.String())
}

func TestCommit_VCS_Fallback_Azure(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "azure")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	hash, _ := InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "", out.String())
}

func TestIdentify_Buildkite(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("BUILDKITE_COMMIT", "abc123")
	_ = os.Setenv("BUILDKITE_PIPELINE_SLUG", "reponame")
	_ = os.Setenv("BUILDKITE_BRANCH_NAME", "feature/first test")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, "abc123", result.Commit())
	assert.Equal(t, "reponame", result.BuildName())
	assert.Equal(t, "feature/first test", result.Branch())
	assert.Equal(t, "feature_first_test", result.BranchReplaceSlash())
	assert.Equal(t, "", out.String())
}

func TestBuildName_Fallback_Buildkite(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "buildkite")
	_ = os.Setenv("BUILDKITE_TOKEN", "abc123")

	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()
	oldPwd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldPwd) }()

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, filepath.Base(dir), result.BuildName())
	assert.Equal(t, "", out.String())
}

func TestBranch_VCS_Fallback_Buildkite(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "buildkite")
	_ = os.Setenv("BUILDKITE_TOKEN", "abc123")

	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, "master", result.Branch())
	assert.Equal(t, "", out.String())
}

func TestCommit_VCS_Fallback_Buildkite(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "buildkite")
	_ = os.Setenv("BUILDKITE_TOKEN", "abc123")

	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	hash, _ := InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "", out.String())
}

func TestIdentify_Gitlab(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature/first test")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, "abc123", result.Commit())
	assert.Equal(t, "reponame", result.BuildName())
	assert.Equal(t, "feature/first test", result.Branch())
	assert.Equal(t, "feature_first_test", result.BranchReplaceSlash())
	assert.Equal(t, "", out.String())
}

func TestName_Gitlab(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature/first test")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.Equal(t, "Gitlab", result.Name())
}

func TestBuildName_Fallback_Gitlab(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "gitlab")

	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()
	oldPwd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldPwd) }()

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, filepath.Base(dir), result.BuildName())
	assert.Equal(t, "", out.String())
}

func TestBranch_VCS_Fallback_Gitlab(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "gitlab")

	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, "master", result.Branch())
	assert.Equal(t, "", out.String())
}

func TestCommit_VCS_Fallback_Gitlab(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "gitlab")

	dir, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	hash, _ := InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "", out.String())
}

func TestNoOp(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "")

	// NoOp uses PWD to generate BuildName so have to switch working dir
	oldPwd, _ := os.Getwd()
	_ = os.Chdir(name)
	defer func() { _ = os.Chdir(oldPwd) }()

	InitRepoWithCommit(name)

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, filepath.Base(name), result.BuildName())
	assert.Equal(t, "master", result.BranchReplaceSlash())
	assert.False(t, result.Configured())
	assert.Equal(t, "", out.String())
}

func TestName_NoOp(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.Equal(t, "none", result.Name())
}

func TestBranch_VCS_Fallback_NoOp(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, "master", result.Branch())
	assert.Equal(t, "", out.String())
}

func TestCommit_VCS_Fallback_NoOp(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	hash, _ := InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NotNil(t, result)
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "", out.String())
}

// Pretty ugly...
func TestCIBuildNameLowerCase(t *testing.T) {
	for _, ci := range initEmptyConfig().availableCI {
		r := reflect.ValueOf(ci).Elem()
		r.FieldByName("CIBuildName").Set(reflect.ValueOf("MixedCase"))
		assert.Equal(t, "mixedcase", ci.BuildName(), "CI %s does not set buildname to lowercase", ci.Name())
	}
}
