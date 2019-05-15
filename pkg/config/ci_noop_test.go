package config

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestNoOp(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	InitRepoWithCommit(dir)

	cfg, err := Load(dir)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.BuildName())
	assert.Equal(t, "master", result.BranchReplaceSlash())
}

func TestBranch_VCS_Fallback_NoOp(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "")

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

func TestCommit_VCS_Fallback_NoOp(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "")

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
