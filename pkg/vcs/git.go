package vcs

import (
	"fmt"
	git2 "gopkg.in/src-d/go-git.v4"
	"io"
	"os"
	"path/filepath"
)

type git struct {
	CommonVCS
	repo *git2.Repository
}

func (v *git) Identify(dir string, out io.Writer) bool {
	if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
		return false
	}
	repo, err := git2.PlainOpen(dir)
	if err != nil {
		_, _ = fmt.Fprintf(out, "Unable to open repository: %s\n", err)
		return false
	}
	v.repo = repo
	ref, err := repo.Head()
	if err != nil {
		_, _ = fmt.Fprintf(out, "Unable to fetch head: %s\n", err)
		return false
	}
	v.CurrentCommit = ref.Hash().String()
	v.CurrentBranch = ref.Name().Short()

	return true
}

func (v *git) Name() string {
	return "Git"
}

var _ VCS = &git{}
