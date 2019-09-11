package config

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/file"
	"os"
	"path/filepath"
	"strings"
)

type BuildkiteCI struct {
	*ci
	CICommit     string `env:"BUILDKITE_COMMIT"`
	CIBuildName  string `env:"BUILDKITE_PIPELINE_SLUG"`
	CIBranchName string `env:"BUILDKITE_BRANCH_NAME"`
}

var _ CI = &BuildkiteCI{}

func (c BuildkiteCI) Name() string {
	return "Buildkite"
}

func (c BuildkiteCI) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.Branch(), "/", "_"), " ", "_")
}

func (c BuildkiteCI) BuildName() string {
	if c.CIBuildName != "" {
		return c.CIBuildName
	}
	return c.ci.BuildName()
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

func (c BuildkiteCI) Scaffold(dir, name, repository string) (*string, error) {
	if err := os.Mkdir(filepath.Join(dir, ".buildkite"), 0777); err != nil {
		return nil, err
	}
	return nil, file.Append(".dockerignore", ".buildkite")
}

func (c BuildkiteCI) Badges() string {
	return ""
}

func (c BuildkiteCI) configured() bool {
	return c.CIBuildName != ""
}
