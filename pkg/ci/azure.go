package ci

import (
	"log"
	"os"
)

type azure struct {
	ci
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
