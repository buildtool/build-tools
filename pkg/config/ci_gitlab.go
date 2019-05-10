package config

import (
	"strings"
)

type GitlabCI struct {
	*ci
	CICommit     string `env:"CI_COMMIT_SHA"`
	CIBuildName  string `env:"CI_PROJECT_NAME"`
	CIBranchName string `env:"CI_COMMIT_REF_NAME"`
}

var _ CI = &GitlabCI{}

func (c GitlabCI) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.Branch(), "/", "_"), " ", "_")
}

func (c GitlabCI) BuildName() string {
	return c.CIBuildName
}

func (c GitlabCI) Branch() string {
	if len(c.CIBranchName) == 0 && c.VCS != nil {
		return c.VCS.Branch()
	}
	return c.CIBranchName
}

func (c GitlabCI) Commit() string {
	if len(c.CICommit) == 0 && c.VCS != nil {
		return c.VCS.Commit()
	}
	return c.CICommit
}
