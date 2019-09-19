package vcs

import (
	"io"
)

type No struct {
	CommonVCS
}

func (v No) Identify(dir string, out io.Writer) bool {
	v.CurrentCommit = ""
	v.CurrentBranch = ""

	return true
}

func (v No) Name() string {
	return "none"
}

var _ VCS = &No{}
