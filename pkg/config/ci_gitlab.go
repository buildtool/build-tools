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

func (c GitlabCI) Name() string {
	return "Gitlab"
}

func (c GitlabCI) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.Branch(), "/", "_"), " ", "_")
}

func (c GitlabCI) BuildName() string {
	if c.CIBuildName != "" {
		return c.CIBuildName
	}
	return c.ci.BuildName()
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

func (c GitlabCI) Scaffold(dir, name, repository string) (*string, error) {
	return nil, nil
}

func (c GitlabCI) Badges() string {
	return ""
}

func (c GitlabCI) configured() bool {
	return c.CIBuildName != ""
}
