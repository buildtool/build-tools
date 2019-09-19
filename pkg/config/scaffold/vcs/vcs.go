package vcs

import "io"

type VCS interface {
	Name() string
	ValidateConfig() error
	Configure()
	Validate(name string) error
	Scaffold(name string) (*RepositoryInfo, error)
	Webhook(name, url string) error
	Clone(dir, name, url string, out io.Writer) error
}

type RepositoryInfo struct {
	SSHURL  string
	HTTPURL string
}
