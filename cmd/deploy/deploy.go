package main

import (
	"os"

	"github.com/apex/log"

	"github.com/buildtool/build-tools/pkg/cli"
	"github.com/buildtool/build-tools/pkg/deploy"
	ver "github.com/buildtool/build-tools/pkg/version"
)

var (
	version              = "dev"
	commit               = "none"
	date                 = "unknown"
	exitFunc             = os.Exit
	handler  log.Handler = cli.New(os.Stdout)
)

func main() {
	log.SetHandler(handler)

	dir, _ := os.Getwd()
	exitFunc(deploy.DoDeploy(dir,
		ver.Info{
			Name:        "deploy",
			Description: "deploys the built image to a Kubernetes cluster",
			Version:     version,
			Commit:      commit,
			Date:        date,
		},
		os.Args[1:]...))
}
