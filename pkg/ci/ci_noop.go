package ci

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"strings"
)

type NoOpCI struct {
	*CommonCI
}

var _ CI = &NoOpCI{}

func (c NoOpCI) Name() string {
	return "none"
}

func (c NoOpCI) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.Branch(), "/", "_"), " ", "_")
}

func (c NoOpCI) BuildName() string {
	return c.CommonCI.BuildName()
}

func (c NoOpCI) Branch() string {
	return c.VCS.Branch()
}

func (c NoOpCI) Commit() string {
	return c.VCS.Commit()
}

func (c NoOpCI) Validate(name string) error {
	return nil
}

func (c NoOpCI) Scaffold(dir string, data templating.TemplateData) (*string, error) {
	return nil, nil
}

func (c NoOpCI) Badges(name string) ([]templating.Badge, error) {
	return nil, nil
}

func (c NoOpCI) Configure() error {
	return nil
}

func (c NoOpCI) Configured() bool {
	return false
}
