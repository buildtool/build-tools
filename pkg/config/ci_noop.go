package config

import (
	"strings"
)

type noOpCI struct {
	*ci
}

var _ CI = &noOpCI{}

func (c noOpCI) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.Branch(), "/", "_"), " ", "_")
}

func (c noOpCI) BuildName() string {
	return ""
}

func (c noOpCI) Branch() string {
	return c.VCS.Branch()
}

func (c noOpCI) Commit() string {
	return c.VCS.Commit()
}
