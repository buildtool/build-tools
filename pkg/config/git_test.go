package config

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestGit_Identify(t *testing.T) {
	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	hash, _ := InitRepoWithCommit(dir)

	result := Identify(dir)
	assert.NotNil(t, result)
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "master", result.Branch())
}

func TestGit_MissingRepo(t *testing.T) {
	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	_ = os.Mkdir(filepath.Join(dir, ".git"), 0777)

	result := Identify(dir)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.Commit())
	assert.Equal(t, "", result.Branch())
}

func TestGit_NoCommit(t *testing.T) {
	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	InitRepo(dir)

	result := Identify(dir)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.Commit())
	assert.Equal(t, "", result.Branch())
}
