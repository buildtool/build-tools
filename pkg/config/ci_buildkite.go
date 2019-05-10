package config

import (
	"strings"
)

type BuildkiteCI struct {
	*ci
	CICommit     string `env:"BUILDKITE_COMMIT"`
	CIBuildName  string `env:"BUILDKITE_PIPELINE_SLUG"`
	CIBranchName string `env:"BUILDKITE_BRANCH_NAME"`
}

var _ CI = &BuildkiteCI{}

func (c BuildkiteCI) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.Branch(), "/", "_"), " ", "_")
}

func (c BuildkiteCI) BuildName() string {
	return c.CIBuildName
}

func (c BuildkiteCI) Branch() string {
	if len(c.CIBranchName) == 0 {
		return c.VCS.Branch()
	}
	return c.CIBranchName
}

func (c BuildkiteCI) Commit() string {
	if len(c.CICommit) == 0 {
		return c.VCS.Commit()
	}
	return c.CICommit
}
