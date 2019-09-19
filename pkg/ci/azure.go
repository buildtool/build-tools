package ci

import (
	"strings"
)

type Azure struct {
	*Common
	CICommit     string `env:"BUILD_SOURCEVERSION"`
	CIBuildName  string `env:"BUILD_REPOSITORY_NAME"`
	CIBranchName string `env:"BUILD_SOURCEBRANCHNAME"`
}

var _ CI = &Azure{}

func (c Azure) Name() string {
	return "Azure"
}

func (c Azure) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.Branch(), "/", "_"), " ", "_")
}

func (c Azure) BuildName() string {
	if c.CIBuildName != "" {
		return strings.ToLower(c.CIBuildName)
	}
	return c.Common.BuildName()
}

func (c Azure) Branch() string {
	if len(c.CIBranchName) == 0 {
		return c.VCS.Branch()
	}
	return c.CIBranchName
}

func (c Azure) Commit() string {
	if len(c.CICommit) == 0 {
		return c.VCS.Commit()
	}
	return c.CICommit
}

func (c Azure) Configured() bool {
	return c.CIBuildName != ""
}
