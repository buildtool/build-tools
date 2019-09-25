package ci

import (
	"strings"
)

type Github struct {
	*Common
	CICommit     string `env:"GITHUB_SHA"`
	CIBuildName  string `env:"RUNNER_WORKSPACE"`
	CIBranchName string `env:"GITHUB_REF"`
}

var _ CI = &Github{}

func (c *Github) Name() string {
	return "Github"
}

func (c *Github) BranchReplaceSlash() string {
	return branchReplaceSlash(c.Branch())
}

func (c *Github) BuildName() string {
	if c.CIBuildName != "" {
		return strings.ToLower(strings.TrimPrefix(c.CIBuildName, "/home/runner/work/"))
	}
	return c.Common.BuildName()
}

func (c *Github) Branch() string {
	if len(c.CIBranchName) == 0 && c.VCS != nil {
		return c.VCS.Branch()
	}
	return strings.TrimPrefix(c.CIBranchName, "refs/head/")
}

func (c *Github) Commit() string {
	if len(c.CICommit) == 0 && c.VCS != nil {
		return c.VCS.Commit()
	}
	return c.CICommit
}

func (c *Github) Configured() bool {
	return c.CIBuildName != ""
}
