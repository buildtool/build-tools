package config

import (
	"os"
	"path/filepath"
)

type CI interface {
	Name() string
	BuildName() string
	Branch() string
	BranchReplaceSlash() string
	Commit() string
	Scaffold(name, repository string) *string
	setVCS(cfg Config)
	configured() bool
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
