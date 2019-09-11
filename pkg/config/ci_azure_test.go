package config

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
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
