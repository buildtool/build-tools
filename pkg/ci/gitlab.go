package ci

import (
	"strings"
)

type Gitlab struct {
	*Common
	CICommit     string `env:"CI_COMMIT_SHA"`
	CIBuildName  string `env:"CI_PROJECT_NAME"`
	CIBranchName string `env:"CI_COMMIT_REF_NAME"`
}

var _ CI = &Gitlab{}

func (c *Gitlab) Name() string {
	return "Gitlab"
}

func (c *Gitlab) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.Branch(), "/", "_"), " ", "_")
}

func (c *Gitlab) BuildName() string {
	if c.CIBuildName != "" {
		return strings.ToLower(c.CIBuildName)
	}
	return c.Common.BuildName()
}

func (c *Gitlab) Branch() string {
	if len(c.CIBranchName) == 0 && c.VCS != nil {
		return c.VCS.Branch()
	}
	return c.CIBranchName
}

func (c *Gitlab) Commit() string {
	if len(c.CICommit) == 0 && c.VCS != nil {
		return c.VCS.Commit()
	}
	return c.CICommit
}

func (c *Gitlab) Configured() bool {
	return c.CIBuildName != ""
}
