package ci

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"strings"
)

type TeamCityCI struct {
	*CommonCI
	CICommit     string `env:"BUILD_VCS_NUMBER"`
	CIBuildName  string `env:"TEAMCITY_PROJECT_NAME"`
	CIBranchName string `env:"BUILD_VCS_BRANCH"`
}

var _ CI = &TeamCityCI{}

func (c TeamCityCI) Name() string {
	return "TeamCityCI"
}

func (c TeamCityCI) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.Branch(), "/", "_"), " ", "_")
}

func (c TeamCityCI) BuildName() string {
	if c.CIBuildName != "" {
		return c.CIBuildName
	}
	return c.CommonCI.BuildName()
}

func (c TeamCityCI) Branch() string {
	if len(c.CIBranchName) == 0 {
		return c.VCS.Branch()
	}
	return c.CIBranchName
}

func (c TeamCityCI) Commit() string {
	if len(c.CICommit) == 0 {
		return c.VCS.Commit()
	}
	return c.CICommit
}

func (c TeamCityCI) Validate(name string) error {
	return nil
}

func (c TeamCityCI) Scaffold(dir string, data templating.TemplateData) (*string, error) {
	return nil, nil
}

func (c TeamCityCI) Badges(name string) ([]templating.Badge, error) {
	return nil, nil
}

func (c TeamCityCI) Configure() error {
	return nil
}

func (c TeamCityCI) Configured() bool {
	return c.CIBuildName != ""
}
