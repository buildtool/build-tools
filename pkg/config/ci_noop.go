package config

import (
	"strings"
)

type noOpCI struct {
	*ci
}

var _ CI = &noOpCI{}

func (c noOpCI) Name() string {
	return "none"
}

func (c noOpCI) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.Branch(), "/", "_"), " ", "_")
}

func (c noOpCI) BuildName() string {
	return c.ci.BuildName()
}

func (c noOpCI) Branch() string {
	return c.VCS.Branch()
}

func (c noOpCI) Commit() string {
	return c.VCS.Commit()
}

func (c noOpCI) Scaffold(name, repository string) *string {
	return nil
}

func (c noOpCI) configured() bool {
	return false
}
