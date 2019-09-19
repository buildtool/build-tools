package ci

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildkite_Name(t *testing.T) {
	ci := &Buildkite{}
	assert.Equal(t, "Buildkite", ci.Name())
}

func TestBuildkite_BuildName(t *testing.T) {
	ci := &Buildkite{CIBuildName: "Name"}

	assert.Equal(t, "name", ci.BuildName())
}

func TestBuildkite_BuildName_Fallback(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	oldpwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldpwd) }()
	_ = os.Chdir(name)

	ci := &Buildkite{Common: &Common{}}

	assert.Equal(t, filepath.Base(name), ci.BuildName())
}

func TestBuildkite_Branch(t *testing.T) {
	ci := &Buildkite{CIBranchName: "feature1"}

	assert.Equal(t, "feature1", ci.Branch())
}

func TestBuildkite_Branch_Fallback(t *testing.T) {
	ci := &Buildkite{Common: &Common{VCS: &mockVcs{}}}

	assert.Equal(t, "fallback-branch", ci.Branch())
}

func TestBuildkite_Commit(t *testing.T) {
	ci := &Buildkite{CICommit: "sha"}

	assert.Equal(t, "sha", ci.Commit())
}

func TestBuildkite_Commit_Fallback(t *testing.T) {
	ci := &Buildkite{Common: &Common{VCS: &mockVcs{}}}

	assert.Equal(t, "fallback-sha", ci.Commit())
}
