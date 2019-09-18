package ci

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"gitlab.com/sparetimecoders/build-tools/pkg/vcs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestTeamCityCI_Name(t *testing.T) {
	ci := &TeamCityCI{}

	assert.Equal(t, "TeamCityCI", ci.Name())
}

func TestTeamCityCI_BranchReplaceSlash(t *testing.T) {
	ci := &TeamCityCI{CIBranchName: "refs/heads/feature1"}

	assert.Equal(t, "refs_heads_feature1", ci.BranchReplaceSlash())
}

func TestTeamCityCI_BranchReplaceSlash_VCS_Fallback(t *testing.T) {
	ci := &TeamCityCI{CommonCI: &CommonCI{VCS: &vcs.Git{CommonVCS: vcs.CommonVCS{CurrentBranch: "refs/heads/feature1"}}}}

	assert.Equal(t, "refs_heads_feature1", ci.BranchReplaceSlash())
}

func TestTeamCityCI_BuildName(t *testing.T) {
	ci := &TeamCityCI{CIBuildName: "project"}

	assert.Equal(t, "project", ci.BuildName())
}

func TestTeamCityCI_BuildName_VCS_Fallback(t *testing.T) {
	oldpwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Chdir(name)
	defer func() { _ = os.Chdir(oldpwd) }()

	ci := &TeamCityCI{CommonCI: &CommonCI{VCS: &vcs.Git{CommonVCS: vcs.CommonVCS{}}}}

	assert.Equal(t, filepath.Base(name), ci.BuildName())
}

func TestTeamCityCI_Commit(t *testing.T) {
	ci := &TeamCityCI{CICommit: "abc123"}

	assert.Equal(t, "abc123", ci.Commit())
}

func TestTeamCityCI_Commit_VCS_Fallback(t *testing.T) {
	ci := &TeamCityCI{CommonCI: &CommonCI{VCS: &vcs.Git{CommonVCS: vcs.CommonVCS{CurrentCommit: "abc123"}}}}

	assert.Equal(t, "abc123", ci.Commit())
}

func TestTeamCityCI_Validate(t *testing.T) {
	ci := &TeamCityCI{}

	assert.NoError(t, ci.Validate("project"))
}

func TestTeamCityCI_Scaffold(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	ci := &TeamCityCI{}

	hookUrl, err := ci.Scaffold(name, templating.TemplateData{})
	assert.NoError(t, err)
	assert.Nil(t, hookUrl)
}

func TestTeamCityCI_Badges(t *testing.T) {
	ci := &TeamCityCI{}

	badges, err := ci.Badges("project")
	assert.NoError(t, err)
	assert.Nil(t, badges)
}

func TestTeamCityCI_Configure(t *testing.T) {
	ci := &TeamCityCI{}

	assert.NoError(t, ci.Configure())
}

func TestTeamCityCI_Configured(t *testing.T) {
	ci := &TeamCityCI{CIBuildName: "project"}

	assert.True(t, ci.Configured())
}
