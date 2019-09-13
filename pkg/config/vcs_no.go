package config

import (
	"io"
	"os"
	"path/filepath"
)

type no struct {
	vcs
}

func (v no) identify(dir string, out io.Writer) bool {
	v.commit = ""
	v.branch = ""

	return true
}

func (v no) configure() {}

func (v no) Name() string {
	return "none"
}

func (v no) Scaffold(name string) (*RepositoryInfo, error) {
	return &RepositoryInfo{}, nil
}

func (v no) Webhook(name, url string) error {
	return nil
}

func (v no) Validate(name string) error {
	return nil
}

func (v no) Clone(dir, name, url string, out io.Writer) error {
	_ = os.MkdirAll(filepath.Join(dir, name), 0777)
	return nil
}

var _ VCS = &no{}
