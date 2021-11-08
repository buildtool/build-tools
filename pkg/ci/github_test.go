package ci

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/buildtool/build-tools/pkg/vcs"
)

func TestGithub_Name(t *testing.T) {
	ci := &Github{}
	assert.Equal(t, "Github", ci.Name())
}

func TestGithub_BuildName(t *testing.T) {
	ci := &Github{Common: &Common{}, CIBuildName: "/home/runner/work/name"}

	assert.Equal(t, "name", ci.BuildName())
}

func TestGithub_Override_ImageName(t *testing.T) {
	ci := &Github{Common: &Common{}, CIBuildName: "/home/runner/work/name"}
	ci.SetImageName("override")
	assert.Equal(t, "override", ci.BuildName())
}

func TestGithub_BuildName_Fallback(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	oldpwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldpwd) }()
	_ = os.Chdir(name)

	ci := &Github{Common: &Common{}}

	assert.Equal(t, filepath.Base(name), ci.BuildName())
}

func TestGithub_BranchReplaceSlash(t *testing.T) {
	ci := &Github{CIBranchName: "refs/heads/feature/xyz"}

	assert.Equal(t, "feature_xyz", ci.BranchReplaceSlash())
}

func TestGithub_Branch_Ref(t *testing.T) {
	ci := &Github{CIBranchName: "refs/heads/feature1"}

	assert.Equal(t, "feature1", ci.Branch())
}

func TestGithub_Branch_Tag(t *testing.T) {
	ci := &Github{CIBranchName: "refs/tags/feature1"}

	assert.Equal(t, "feature1", ci.Branch())
}
func TestGithub_Branch(t *testing.T) {
	ci := &Github{CIBranchName: "feature1"}

	assert.Equal(t, "feature1", ci.Branch())
}

func TestGithub_Branch_Fallback(t *testing.T) {
	ci := &Github{Common: &Common{VCS: vcs.NewMockVcs()}}

	assert.Equal(t, "fallback-branch", ci.Branch())
}

func TestGithub_Commit(t *testing.T) {
	ci := &Github{CICommit: "sha"}

	assert.Equal(t, "sha", ci.Commit())
}

func TestGithub_Commit_Fallback(t *testing.T) {
	ci := &Github{Common: &Common{VCS: vcs.NewMockVcs()}}

	assert.Equal(t, "fallback-sha", ci.Commit())
}
