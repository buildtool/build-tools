//go:build !prod
// +build !prod

package config

import (
	"io/ioutil"

	git2 "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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
