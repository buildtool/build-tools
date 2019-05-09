package ci

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/vcs"
	"log"
	"os"
)

type gitlab struct {
	ci
}

var _ CI = &gitlab{}

func (g *gitlab) identify(vcs vcs.VCS) bool {
	if _, exists := os.LookupEnv("GITLAB_CI"); exists {
		log.Println("Running in GitlabCI")
		g.CICommit = os.Getenv("CI_COMMIT_SHA")
		g.CIBuildName = os.Getenv("CI_PROJECT_NAME")
		g.CIBranchName = os.Getenv("CI_COMMIT_REF_NAME")
		g.VCS = vcs
		return true
	}
	return false
}
