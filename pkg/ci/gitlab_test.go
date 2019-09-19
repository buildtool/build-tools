package ci

import (
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
