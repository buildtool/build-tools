package ci

import (
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/vcs"
	"strings"
)

type CI interface {
	identify(vcs vcs.VCS) bool
	// TODO: Uncomment when implementing service-setup
	//Validate() bool
	//Scaffold() error
	BuildName() string
	Branch() string
	BranchReplaceSlash() string
	Commit() string
}

var cis = []CI{&azure{}, &buildkite{}, &gitlab{}}

func Identify(vcs vcs.VCS) (CI, error) {
	for _, ci := range cis {
		if ci.identify(vcs) {
			return ci, nil
		}
	}

	return nil, fmt.Errorf("no CI found")
}

type ci struct {
	CICommit     string
	CIBuildName  string
	CIBranchName string
	VCS          vcs.VCS
}

func (c ci) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.Branch(), "/", "_"), " ", "_")
}

func (c ci) BuildName() string {
	return c.CIBuildName
}

func (c ci) Branch() string {
	if len(c.CIBranchName) == 0 {
		return c.VCS.Branch()
	}
	return c.CIBranchName
}

func (c ci) Commit() string {
	if len(c.CICommit) == 0 {
		return c.VCS.Commit()
	}
	return c.CICommit
}
