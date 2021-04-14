package main

import (
	"os"

	"github.com/mattn/go-colorable"

	"github.com/buildtool/build-tools/pkg/deploy"
	ver "github.com/buildtool/build-tools/pkg/version"
)

var (
	version  = "dev"
	commit   = "none"
	date     = "unknown"
	exitFunc = os.Exit
	out      = colorable.NewColorableStdout()
	eout     = colorable.NewColorableStderr()
)

func main() {
	dir, _ := os.Getwd()

	exitFunc(deploy.DoDeploy(dir, out, eout,
		ver.Info{
			Name:        "deploy",
			Description: "deploys the built image to a Kubernetes cluster",
			Version:     version,
			Commit:      commit,
			Date:        date,
		},
		os.Args[1:]...))
}
