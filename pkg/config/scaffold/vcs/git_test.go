package vcs

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	git2 "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io/ioutil"
	"os"
	"testing"
)

func TestGit_Clone(t *testing.T) {
	vcs := &Git{}

	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	dir, _ := ioutil.TempDir(os.TempDir(), "Git-repo")
	defer func() { _ = os.RemoveAll(dir) }()
	repo, _ := git2.PlainInit(dir, false)
	tree, _ := repo.Worktree()
	file, _ := ioutil.TempFile(dir, "file")
	_, _ = file.WriteString("test")
	_ = file.Close()
	_, _ = tree.Add(file.Name())
	_, _ = tree.Commit("Test", &git2.CommitOptions{Author: &object.Signature{Email: "test@example.com"}})

	buff := &bytes.Buffer{}
	err := vcs.Clone(name, "project", fmt.Sprintf("file://%s", dir), buff)
	assert.NoError(t, err)
	assert.Contains(t, buff.String(), "Total 2 (delta 0), reused 0 (delta 0)")
}
