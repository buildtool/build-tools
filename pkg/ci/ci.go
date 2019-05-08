package ci

import (
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

func Identify() CI {
  for _, ci := range cis {
    if ci.identify() {
      return ci
    }
  }

  return nil
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
