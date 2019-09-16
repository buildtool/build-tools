package config

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

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
	result, err := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, filepath.Base(name), result.BuildName())
	assert.Equal(t, "master", result.BranchReplaceSlash())
	assert.False(t, result.configured())
	assert.Equal(t, "", out.String())
}

func TestName_NoOp(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI", "")

	out := &bytes.Buffer{}
	cfg, err := Load(name, out)
	assert.NoError(t, err)
	result, err := cfg.CurrentCI()
	assert.NoError(t, err)
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
	result, err := cfg.CurrentCI()
	assert.NoError(t, err)
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
	result, err := cfg.CurrentCI()
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "", out.String())
}

func TestNoOpCI_Scaffold(t *testing.T) {
	ci := &noOpCI{}

	hookUrl, err := ci.Scaffold(name, templating.TemplateData{})
	assert.Nil(t, hookUrl)
	assert.NoError(t, err)
}

func TestNoOpCI_Badges(t *testing.T) {
	ci := &noOpCI{}

	badges, err := ci.Badges(name)
	assert.Nil(t, badges)
	assert.NoError(t, err)
}
