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

func (c noOpCI) Scaffold(dir, name, repository string) (*string, error) {
	return nil, nil
}

func (c noOpCI) Badges() string {
	return ""
}

func (c noOpCI) configured() bool {
	return false
}
