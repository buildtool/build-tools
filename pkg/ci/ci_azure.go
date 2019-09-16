package ci

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"strings"
)

type AzureCI struct {
	*CommonCI
	CICommit     string `env:"BUILD_SOURCEVERSION"`
	CIBuildName  string `env:"BUILD_REPOSITORY_NAME"`
	CIBranchName string `env:"BUILD_SOURCEBRANCHNAME"`
}

var _ CI = &AzureCI{}

func (c AzureCI) Name() string {
	return "Azure"
}

func (c AzureCI) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.Branch(), "/", "_"), " ", "_")
}

func (c AzureCI) BuildName() string {
	if c.CIBuildName != "" {
		return c.CIBuildName
	}
	return c.CommonCI.BuildName()
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

func (c AzureCI) Validate(name string) error {
	return nil
}

func (c AzureCI) Scaffold(dir string, data templating.TemplateData) (*string, error) {
	return nil, nil
}

func (c AzureCI) Badges(name string) ([]templating.Badge, error) {
	return nil, nil
}

func (c AzureCI) Configure() error {
	return nil
}

func (c AzureCI) Configured() bool {
	return c.CIBuildName != ""
}
