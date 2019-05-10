// +build !prod

package config

import (
	git2 "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io/ioutil"
)

func InitRepo(dir string) *git2.Repository {
	repo, _ := git2.PlainInit(dir, false)
	return repo
}
func InitRepoWithCommit(dir string) (plumbing.Hash, *git2.Repository) {
	repo := InitRepo(dir)
	tree, _ := repo.Worktree()
	file, _ := ioutil.TempFile(dir, "file")
	_, _ = file.WriteString("test")
	_ = file.Close()
	_, _ = tree.Add(file.Name())
	hash, _ := tree.Commit("Test", &git2.CommitOptions{Author: &object.Signature{Email: "test@example.com"}})

	return hash, repo
}
