package main

import (
	"os"

	"github.com/apex/log"

	"github.com/buildtool/build-tools/pkg/cli"
	"github.com/buildtool/build-tools/pkg/prepare"
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
	exitFunc(prepare.DoPrepare(dir,
		ver.Info{
			Name:        "build",
			Description: "performs a docker build and tags the resulting image",
			Version:     version,
			Commit:      commit,
			Date:        date,
		},
		os.Args[1:]...))
}
