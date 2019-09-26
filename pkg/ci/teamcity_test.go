package ci

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/vcs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestTeamCityCI_Name(t *testing.T) {
	ci := &TeamCity{}

	assert.Equal(t, "TeamCity", ci.Name())
}

func TestTeamCityCI_BranchReplaceSlash(t *testing.T) {
	ci := &TeamCity{CIBranchName: "refs/heads/feature1"}

	assert.Equal(t, "refs_heads_feature1", ci.BranchReplaceSlash())
}

func TestTeamCityCI_BranchReplaceSlash_VCS_Fallback(t *testing.T) {
	ci := &TeamCity{Common: &Common{VCS: vcs.NewMockVcsWithBranch( "refs/heads/feature1")}}

	assert.Equal(t, "refs_heads_feature1", ci.BranchReplaceSlash())
}

func TestTeamCityCI_BuildName(t *testing.T) {
	ci := &TeamCity{CIBuildName: "project"}

	assert.Equal(t, "project", ci.BuildName())
}

func TestTeamCityCI_BuildName_VCS_Fallback(t *testing.T) {
	oldpwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Chdir(name)
	defer func() { _ = os.Chdir(oldpwd) }()

	ci := &TeamCity{Common: &Common{VCS: vcs.NewMockVcs()}}

	assert.Equal(t, filepath.Base(name), ci.BuildName())
}

func TestTeamCityCI_Commit(t *testing.T) {
	ci := &TeamCity{CICommit: "abc123"}

	assert.Equal(t, "abc123", ci.Commit())
}

func TestTeamCityCI_Commit_VCS_Fallback(t *testing.T) {
	ci := &TeamCity{Common: &Common{VCS: vcs.NewMockVcs()}}

	assert.Equal(t, "fallback-sha", ci.Commit())
}

func TestTeamCityCI_Configured(t *testing.T) {
	ci := &TeamCity{CIBuildName: "project"}

	assert.True(t, ci.Configured())
}
