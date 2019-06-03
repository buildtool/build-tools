package config

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestIdentify_Gitlab(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("GITLAB_CI", "1")
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "reponame")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "feature/first test")

	cfg, err := Load(".")
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "abc123", result.Commit())
	assert.Equal(t, "reponame", result.BuildName())
	assert.Equal(t, "feature/first test", result.Branch())
	assert.Equal(t, "feature_first_test", result.BranchReplaceSlash())
}

func TestBuildName_Fallback_Gitlab(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "gitlab")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)
	_ = os.Chdir(dir)

	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, filepath.Base(dir), result.BuildName())
}

func TestBranch_VCS_Fallback_Gitlab(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "gitlab")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	InitRepoWithCommit(dir)

	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "master", result.Branch())
}

func TestCommit_VCS_Fallback_Gitlab(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "gitlab")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	hash, _ := InitRepoWithCommit(dir)

	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, hash.String(), result.Commit())
}
