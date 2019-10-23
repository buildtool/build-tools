package ci

import (
	"github.com/sparetimecoders/build-tools/pkg/vcs"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGitlab_Name(t *testing.T) {
	ci := &Gitlab{}
	assert.Equal(t, "Gitlab", ci.Name())
}

func TestGitlab_BuildName(t *testing.T) {
	ci := &Gitlab{CIBuildName: "Name"}

	assert.Equal(t, "name", ci.BuildName())
}

func TestGitlab_BuildName_Fallback(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	oldpwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldpwd) }()
	_ = os.Chdir(name)

	ci := &Gitlab{Common: &Common{}}

	assert.Equal(t, filepath.Base(name), ci.BuildName())
}

func TestGitlab_Branch(t *testing.T) {
	ci := &Gitlab{CIBranchName: "feature1"}

	assert.Equal(t, "feature1", ci.Branch())
}

func TestGitlab_Branch_Fallback(t *testing.T) {
	ci := &Gitlab{Common: &Common{VCS: vcs.NewMockVcs()}}

	assert.Equal(t, "fallback-branch", ci.Branch())
}

func TestGitlab_Commit(t *testing.T) {
	ci := &Gitlab{CICommit: "sha"}

	assert.Equal(t, "sha", ci.Commit())
}

func TestGitlab_Commit_Fallback(t *testing.T) {
	ci := &Gitlab{Common: &Common{VCS: vcs.NewMockVcs()}}

	assert.Equal(t, "fallback-sha", ci.Commit())
}
