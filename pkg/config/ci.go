package config

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/templating"
	"os"
	"path/filepath"
)

type CI interface {
	Name() string
	BuildName() string
	Branch() string
	BranchReplaceSlash() string
	Commit() string
	Scaffold(dir string, data templating.TemplateData) (*string, error)
	Badges(name string) ([]templating.Badge, error)
	setVCS(cfg Config)
	configured() bool
	configure()
}

type ci struct {
	VCS VCS
}

func (c *ci) setVCS(cfg Config) {
	c.VCS = cfg.CurrentVCS()
}

func (c *ci) BuildName() string {
	dir, _ := os.Getwd()
	return filepath.Base(dir)
}
