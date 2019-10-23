package ci

import (
	"github.com/sparetimecoders/build-tools/pkg/vcs"
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

func (c *Common) BuildName(name string) string {
	if name != "" {
		return strings.ToLower(name)
	}
	dir, _ := os.Getwd()
	return strings.ToLower(filepath.Base(dir))
}

func (c *Common) Branch(name string) string {
	if name != "" {
		return name
	}
	return c.VCS.Branch()
}

func (c *Common) Commit(name string) string {
	if name != "" {
		return name
	}
	return c.VCS.Commit()
}

func branchReplaceSlash(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(name, "/", "_"), " ", "_")
}

func IsValid(c CI) bool {
	return len(c.Commit()) != 0 || len(c.Branch()) != 0
}
