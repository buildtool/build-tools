package config

import (
	"io"
)

type VCS interface {
	identify(dir string, out io.Writer) bool
	Name() string
	Branch() string
	Commit() string
	Scaffold(name string) (*RepositoryInfo, error)
	Webhook(name, url string) error
	Clone(dir, name, url string, out io.Writer) error
	Validate() error
}

type RepositoryInfo struct {
	SSHURL  string
	HTTPURL string
}

var systems = []VCS{&git{}}

func Identify(dir string, out io.Writer) VCS {
	for _, vcs := range systems {
		if vcs.identify(dir, out) {
			return vcs
		}
	}

	no := &no{}
	no.identify(dir, out)
	return no
}

type vcs struct {
	branch string
	commit string
}

func (v vcs) Branch() string {
	return v.branch
}

func (v vcs) Commit() string {
	return v.commit
}
