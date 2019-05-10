package config

import (
	"strings"
)

type AzureCI struct {
	*ci
	CICommit     string `env:"BUILD_SOURCEVERSION"`
	CIBuildName  string `env:"BUILD_REPOSITORY_NAME"`
	CIBranchName string `env:"BUILD_SOURCEBRANCHNAME"`
}

var _ CI = &AzureCI{}

func (c AzureCI) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.Branch(), "/", "_"), " ", "_")
}

func (c AzureCI) BuildName() string {
	return c.CIBuildName
}

func (c AzureCI) Branch() string {
	if len(c.CIBranchName) == 0 {
		return c.VCS.Branch()
	}
	return c.CIBranchName
}

func (c AzureCI) Commit() string {
	if len(c.CICommit) == 0 {
		return c.VCS.Commit()
	}
	return c.CICommit
}
