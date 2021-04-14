package main

import (
	"os"

	"github.com/mattn/go-colorable"

	"github.com/buildtool/build-tools/pkg/build"
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
	exitFunc(build.DoBuild(dir, out, eout,
		ver.Info{
			Name:        "build",
			Description: "performs a docker build and tags the resulting image",
			Version:     version,
			Commit:      commit,
			Date:        date,
		},
		os.Args[1:]...))
}
