package vcs

import (
	"io"
)

type no struct {
	CommonVCS
}

func (v no) Identify(dir string, out io.Writer) bool {
	return true
}

func (v no) Name() string {
	return "none"
}

var _ VCS = &no{}
