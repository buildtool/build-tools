package ci

import (
	"log"
	"os"
)

type gitlab struct {
	ci
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
