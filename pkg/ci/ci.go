package ci

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/vcs"
	"os"
	"path/filepath"
	"strings"
)

type CI interface {
	Name() string
	// BuildName returns the name of the current build in lowercase
	BuildName() string
	Branch() string
	BranchReplaceSlash() string
	Commit() string
	SetVCS(vcs vcs.VCS)
	Configured() bool
}

type Common struct {
	VCS vcs.VCS
}

func (c *Common) SetVCS(vcs vcs.VCS) {
	c.VCS = vcs
}

func (c *Common) BuildName() string {
	dir, _ := os.Getwd()
	return strings.ToLower(filepath.Base(dir))
}

func branchReplaceSlash(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(name, "/", "_"), " ", "_")
}
