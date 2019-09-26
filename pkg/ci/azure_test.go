package ci

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/vcs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestAzure_Name(t *testing.T) {
	ci := &Azure{}

	assert.Equal(t, "Azure", ci.Name())
}

func TestAzure_BuildName(t *testing.T) {
	ci := &Azure{CIBuildName: "Name"}

	assert.Equal(t, "name", ci.BuildName())
}

func TestAzure_BuildName_Fallback(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	oldpwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldpwd) }()
	_ = os.Chdir(name)

	ci := &Azure{Common: &Common{}}

	assert.Equal(t, filepath.Base(name), ci.BuildName())
}

func TestAzure_Branch(t *testing.T) {
	ci := &Azure{CIBranchName: "feature1"}

	assert.Equal(t, "feature1", ci.Branch())
}

func TestAzure_Branch_Fallback(t *testing.T) {
	ci := &Azure{Common: &Common{VCS: vcs.NewMockVcs()}}

	assert.Equal(t, "fallback-branch", ci.Branch())
}

func TestAzure_Commit(t *testing.T) {
	ci := &Azure{CICommit: "sha"}

	assert.Equal(t, "sha", ci.Commit())
}

func TestAzure_Commit_Fallback(t *testing.T) {
	ci := &Azure{Common: &Common{VCS: vcs.NewMockVcs()}}

	assert.Equal(t, "fallback-sha", ci.Commit())
}
