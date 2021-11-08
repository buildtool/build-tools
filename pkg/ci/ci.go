package ci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/buildtool/build-tools/pkg/vcs"
)

type CI interface {
	Name() string
	// BuildName returns the name of the current build in lowercase
	BuildName() string
	Branch() string
	BranchReplaceSlash() string
	Commit() string
	SetVCS(vcs vcs.VCS)
	SetImageName(imageName string)
	Configured() bool
}

type Common struct {
	VCS       vcs.VCS
	ImageName string
}

func (c *Common) SetVCS(vcs vcs.VCS) {
	c.VCS = vcs
}

func (c *Common) SetImageName(imageName string) {
	c.ImageName = imageName
}

func (c *Common) BuildName(name string) string {
	if c.ImageName != "" {
		fmt.Printf("Using %s as BuildName", name)
		return c.ImageName
	}
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
