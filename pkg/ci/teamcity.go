package ci

import (
	"strings"
)

type TeamCity struct {
	*Common
	CICommit     string `env:"BUILD_VCS_NUMBER"`
	CIBuildName  string `env:"TEAMCITY_PROJECT_NAME"`
	CIBranchName string `env:"BUILD_VCS_BRANCH"`
}

var _ CI = &TeamCity{}

func (c TeamCity) Name() string {
	return "TeamCity"
}

func (c TeamCity) BranchReplaceSlash() string {
	return branchReplaceSlash(c.Branch())
}

func (c TeamCity) BuildName() string {
	if c.CIBuildName != "" {
		return strings.ToLower(c.CIBuildName)
	}
	return c.Common.BuildName()
}

func (c TeamCity) Branch() string {
	if len(c.CIBranchName) == 0 {
		return c.VCS.Branch()
	}
	return c.CIBranchName
}

func (c TeamCity) Commit() string {
	if len(c.CICommit) == 0 {
		return c.VCS.Commit()
	}
	return c.CICommit
}

func (c TeamCity) Configured() bool {
	return c.CIBuildName != ""
}
