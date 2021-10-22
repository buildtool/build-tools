package main

import (
	"os"

	"github.com/apex/log"

	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/build"
	"github.com/buildtool/build-tools/pkg/cli"
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

	var buildArgs build.Args
	info := ver.Info{
		Name:        "build",
		Description: "performs a docker build and tags the resulting image",
		Version:     version,
		Commit:      commit,
		Date:        date,
	}
	err := args.ParseArgs(dir, os.Args[1:], info, &buildArgs)
	if err != nil {
		if err != args.Done {
			exitFunc(-1)
			return
		} else {
			exitFunc(0)
			return
		}
	}

	if err := build.DoBuild(dir, buildArgs); err != nil {
		log.Error(err.Error())
		exitFunc(-1)
		return
	}
	exitFunc(0)
}
