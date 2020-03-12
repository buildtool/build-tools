package main

import (
	"github.com/buildtool/build-tools/pkg/build"
	ver "github.com/buildtool/build-tools/pkg/version"
	"github.com/mattn/go-colorable"
	"os"
)

var (
	version  = "dev"
	commit   = "none"
	date     = "unknown"
	exitFunc = os.Exit
	out      = colorable.NewColorableStdout()
)

func main() {
	if ver.PrintVersionOnly(version, commit, date, out) {
		exitFunc(0)
	} else {
		dir, _ := os.Getwd()
		exitFunc(build.DoBuild(dir, out, colorable.NewColorableStderr(), os.Args[1:]...))
	}
}
