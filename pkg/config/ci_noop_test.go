package config

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestNoOp(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)
	oldPwd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer os.Chdir(oldPwd)

	InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(".", out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, filepath.Base(dir), result.BuildName())
	assert.Equal(t, "master", result.BranchReplaceSlash())
	assert.False(t, result.configured())
	assert.Equal(t, "", out.String())
}

func TestBranch_VCS_Fallback_NoOp(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "master", result.Branch())
	assert.Equal(t, "", out.String())
}

func TestCommit_VCS_Fallback_NoOp(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "")

	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	hash, _ := InitRepoWithCommit(dir)

	out := &bytes.Buffer{}
	cfg, err := Load(dir, out)
	assert.NoError(t, err)
	result := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "", out.String())
}
