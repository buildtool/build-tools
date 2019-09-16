package vcs

import (
	"fmt"
	git2 "gopkg.in/src-d/go-git.v4"
	"io"
	"os"
	"path/filepath"
)

type Git struct {
	CommonVCS
	repo *git2.Repository
}

func (v *Git) Identify(dir string, out io.Writer) bool {
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

func (v *Git) Configure() {}

func (v *Git) Name() string {
	return "Git"
}

func (v *Git) Scaffold(name string) (*RepositoryInfo, error) {
	return nil, nil
}

func (v *Git) Webhook(name, url string) error {
	return nil
}

func (v *Git) Validate(name string) error {
	return nil
}

func (v *Git) Clone(dir, name, url string, out io.Writer) error {
	_, err := git2.PlainClone(filepath.Join(dir, name), false, &git2.CloneOptions{URL: url, Progress: out})
	return err
}

var _ VCS = &Git{}
