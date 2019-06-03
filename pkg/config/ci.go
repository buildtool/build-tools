package config

import (
	"os"
	"path/filepath"
)

type CI interface {
	// TODO: Uncomment when implementing service-setup
	//Validate() bool
	//Scaffold() error
	BuildName() string
	Branch() string
	BranchReplaceSlash() string
	Commit() string
	setVCS(cfg Config)
	configured() bool
}

type ci struct {
	VCS VCS
}

func (c *ci) setVCS(cfg Config) {
	if v, e := cfg.CurrentVCS(); e == nil {
		c.VCS = v
	}
}

func (c *ci) BuildName() string {
	dir, _ := os.Getwd()
	return filepath.Base(dir)
}
