package vcs

import (
	"fmt"
	git2 "gopkg.in/src-d/go-git.v4"
	"os"
	"path"
)

type git struct {
	vcs
	repo *git2.Repository
}

func (v *git) identify(dir string) bool {
	if _, err := os.Stat(path.Join(dir, ".git")); os.IsNotExist(err) {
		return false
	}
	repo, err := git2.PlainOpen(dir)
	if err != nil {
		fmt.Printf("Unable to open repository: %s\n", err)
		return false
	}
	v.repo = repo
	ref, err := repo.Head()
	if err != nil {
		fmt.Printf("Unable to fetch head: %s\n", err)
		return false
	}
	v.commit = ref.Hash().String()
	v.branch = ref.Name().Short()

	return true
}
