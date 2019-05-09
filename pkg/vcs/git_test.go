package vcs

import (
	"github.com/stretchr/testify/assert"
	git2 "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestGit_Identify(t *testing.T) {
	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	repo, _ := git2.PlainInit(dir, false)
	tree, _ := repo.Worktree()
	file, _ := ioutil.TempFile(dir, "file")
	_, _ = file.WriteString("test")
	_ = file.Close()
	_, _ = tree.Add(file.Name())
	hash, _ := tree.Commit("Test", &git2.CommitOptions{Author: &object.Signature{Email: "test@example.com"}})

	result := Identify(dir)
	assert.NotNil(t, result)
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "master", result.Branch())
}

func TestGit_MissingRepo(t *testing.T) {
	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	_ = os.Mkdir(path.Join(dir, ".git"), 0777)

	result := Identify(dir)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.Commit())
	assert.Equal(t, "", result.Branch())
}

func TestGit_NoCommit(t *testing.T) {
	dir, _ := ioutil.TempDir("", "build-tools")
	defer os.RemoveAll(dir)

	_, _ = git2.PlainInit(dir, false)

	result := Identify(dir)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.Commit())
	assert.Equal(t, "", result.Branch())
}
