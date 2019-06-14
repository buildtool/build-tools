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
