package ci

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/vcs"
	"log"
	"os"
)

type buildkite struct {
	ci
}

var _ CI = &buildkite{}

func (b *buildkite) identify(vcs vcs.VCS) bool {
	if _, exists := os.LookupEnv("BUILDKITE_COMMIT"); exists {
		log.Println("Running in Buildkite")
		b.CICommit = os.Getenv("BUILDKITE_COMMIT")
		b.CIBuildName = os.Getenv("BUILDKITE_PIPELINE_SLUG")
		b.CIBranchName = os.Getenv("BUILDKITE_BRANCH_NAME")
		b.VCS = vcs
		return true
	}
	return false
}
