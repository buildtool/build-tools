package main

import (
	"os"

	"github.com/apex/log"

	"github.com/buildtool/build-tools/pkg/cli"
	"github.com/buildtool/build-tools/pkg/push"
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
	exitFunc(push.Push(dir, ver.Info{
		Name:        "push",
		Description: "push a docker image to a remote docker repository",
		Version:     version,
		Commit:      commit,
		Date:        date,
	},
		os.Args[1:]...))
}
