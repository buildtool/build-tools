package vcs

import (
	"io"
	"os"
	"path/filepath"
)

type No struct {
	CommonVCS
}

func (v No) Identify(dir string, out io.Writer) bool {
	v.CurrentCommit = ""
	v.CurrentBranch = ""

	return true
}

func (v No) Configure() {}

func (v No) Name() string {
	return "none"
}

func (v No) Scaffold(name string) (*RepositoryInfo, error) {
	return &RepositoryInfo{}, nil
}

func (v No) Webhook(name, url string) error {
	return nil
}

func (v No) Validate(name string) error {
	return nil
}

func (v No) Clone(dir, name, url string, out io.Writer) error {
	_ = os.MkdirAll(filepath.Join(dir, name), 0777)
	return nil
}

var _ VCS = &No{}
