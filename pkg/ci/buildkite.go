package ci

import (
	"strings"
)

type Buildkite struct {
	*Common
	CICommit     string `env:"BUILDKITE_COMMIT"`
	CIBuildName  string `env:"BUILDKITE_PIPELINE_SLUG"`
	CIBranchName string `env:"BUILDKITE_BRANCH_NAME"`
}

var _ CI = &Buildkite{}

func (c *Buildkite) Name() string {
	return "Buildkite"
}

func (c *Buildkite) BranchReplaceSlash() string {
	return branchReplaceSlash(c.Branch())
}

func (c *Buildkite) BuildName() string {
	if c.CIBuildName != "" {
		return strings.ToLower(c.CIBuildName)
	}
	return c.Common.BuildName()
}

func (c *Buildkite) Branch() string {
	if len(c.CIBranchName) == 0 {
		return c.VCS.Branch()
	}
	return c.CIBranchName
}

func (c *Buildkite) Commit() string {
	if len(c.CICommit) == 0 {
		return c.VCS.Commit()
	}
	return c.CICommit
}

func (c *Buildkite) Configured() bool {
	return c.CIBuildName != ""
}
