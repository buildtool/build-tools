package config

import (
	"io"
)

type no struct {
	vcs
}

func (v no) identify(dir string, out io.Writer) bool {
	v.commit = ""
	v.branch = ""

	return true
}

func (v no) Name() string {
	return "none"
}

func (v no) Scaffold(name string) string {
	return ""
}

func (v no) Webhook(name, url string) {
}

var _ VCS = &no{}
