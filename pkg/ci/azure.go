package ci

import (
  "log"
  "os"
  "strings"
)

type azure struct {
  CICommit     string
  CIBuildName  string
  CIBranchName string
}

var _ CI = &azure{}

func (a *azure) identify() bool {
  if _, exists := os.LookupEnv("VSTS_PROCESS_LOOKUP_ID"); exists {
    log.Println("Running in Azure")
    a.CICommit = os.Getenv("BUILD_SOURCEVERSION")
    a.CIBuildName = os.Getenv("BUILD_REPOSITORY_NAME")
    a.CIBranchName = os.Getenv("BUILD_SOURCEBRANCHNAME")
    return true
  }
  return false
}

func (a azure) BuildName() string {
  return a.CIBuildName
}

func (a azure) Branch() string {
  return a.CIBranchName
}

func (a azure) BranchReplaceSlash() string {
  return strings.ReplaceAll(strings.ReplaceAll(a.CIBranchName, "/", "_"), " ", "_")
}

func (a azure) Commit() string {
  return a.CICommit
}
