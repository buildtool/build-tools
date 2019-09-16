package vcs

import "io"

type VCS interface {
	Identify(dir string, out io.Writer) bool
	Configure()
	Name() string
	Branch() string
	Commit() string
	Scaffold(name string) (*RepositoryInfo, error)
	Webhook(name, url string) error
	Clone(dir, name, url string, out io.Writer) error
	Validate(name string) error
}

type RepositoryInfo struct {
	SSHURL  string
	HTTPURL string
}

type CommonVCS struct {
	CurrentBranch string
	CurrentCommit string
}

func (v CommonVCS) Branch() string {
	return v.CurrentBranch
}

func (v CommonVCS) Commit() string {
	return v.CurrentCommit
}

var systems = []VCS{&Git{}}

func Identify(dir string, out io.Writer) VCS {
	for _, vcs := range systems {
		if vcs.Identify(dir, out) {
			return vcs
		}
	}

	no := &No{}
	no.Identify(dir, out)
	return no
}
