package ci

import (
	"fmt"
	"strings"
)

type CI interface {
	identify() bool
	// TODO: Uncomment when implementing service-setup
	//Validate() bool
	//Scaffold() error
	BuildName() string
	Branch() string
	BranchReplaceSlash() string
	Commit() string
}

var cis = []CI{&azure{}, &buildkite{}, &gitlab{}}

func Identify() (CI, error) {
	for _, ci := range cis {
		if ci.identify() {
			return ci, nil
		}
	}

	return nil, fmt.Errorf("no CI found")
}

type ci struct {
	CICommit     string
	CIBuildName  string
	CIBranchName string
}

func (c ci) BranchReplaceSlash() string {
	return strings.ReplaceAll(strings.ReplaceAll(c.CIBranchName, "/", "_"), " ", "_")
}

func (c ci) BuildName() string {
	return c.CIBuildName
}

func (c ci) Branch() string {
	return c.CIBranchName
}

func (c ci) Commit() string {
	return c.CICommit
}
