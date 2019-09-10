package config

import (
	"fmt"
	git2 "gopkg.in/src-d/go-git.v4"
	"io"
	"os"
	"path/filepath"
)

type git struct {
	vcs
	repo *git2.Repository
}

func (v *git) identify(dir string, out io.Writer) bool {
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
	v.commit = ref.Hash().String()
	v.branch = ref.Name().Short()

	return true
}

func (v *git) Name() string {
	return "git"
}

func (v *git) Scaffold(name string) (string, error) {
	return "", nil
}

func (v *git) Webhook(name, url string) {
}

func (v *git) Clone(name, url string, out io.Writer) error {
	_, err := git2.PlainClone(name, false, &git2.CloneOptions{URL: url, Progress: out})
	return err
}

var _ VCS = &git{}
