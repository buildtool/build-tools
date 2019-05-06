package ci

import (
  "log"
  "os"
  "strings"
)

type buildkite struct {
  CICommit     string
  CIBuildName  string
  CIBranchName string
}

var _ CI = &buildkite{}

func (b *buildkite) identify() bool {
  if _, exists := os.LookupEnv("BUILDKITE_COMMIT"); exists {
    log.Println("Running in Buildkite")
    b.CICommit = os.Getenv("BUILDKITE_COMMIT")
    b.CIBuildName = os.Getenv("BUILDKITE_PIPELINE_SLUG")
    b.CIBranchName = os.Getenv("BUILDKITE_BRANCH_NAME")
    return true
  }
  return false
}

func (b buildkite) Validate() bool {
  panic("implement me")
}

func (b buildkite) Scaffold() error {
  panic("implement me")
}

func (b buildkite) BuildName() string {
  return b.CIBuildName
}

func (b buildkite) Branch() string {
  return b.CIBranchName
}

func (b buildkite) BranchReplaceSlash() string {
  return strings.ReplaceAll(strings.ReplaceAll(b.CIBranchName, "/", "_"), " ", "_")
}

func (b buildkite) Commit() string {
  return b.CICommit
}
