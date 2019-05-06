package ci

import (
  "log"
  "os"
  "strings"
)

type gitlab struct {
  CICommit     string
  CIBuildName  string
  CIBranchName string
}

var _ CI = &gitlab{}

func (g *gitlab) identify() bool {
  if _, exists := os.LookupEnv("GITLAB_CI"); exists {
    log.Println("Running in GitlabCI")
    g.CICommit = os.Getenv("CI_COMMIT_SHA")
    g.CIBuildName = os.Getenv("CI_PROJECT_NAME")
    g.CIBranchName = os.Getenv("CI_COMMIT_REF_NAME")
    return true
  }
  return false
}

func (g gitlab) Validate() bool {
  panic("implement me")
}

func (g gitlab) Scaffold() error {
  panic("implement me")
}

func (g gitlab) BuildName() string {
  return g.CIBuildName
}

func (g gitlab) Branch() string {
  return g.CIBranchName
}

func (g gitlab) BranchReplaceSlash() string {
  return strings.ReplaceAll(strings.ReplaceAll(g.CIBranchName, "/", "_"), " ", "_")
}

func (g gitlab) Commit() string {
  return g.CICommit
}
