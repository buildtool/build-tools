package main

import (
	"os"

	"github.com/mattn/go-colorable"

	"github.com/buildtool/build-tools/pkg/push"
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
	exitFunc(push.Push(dir, out, eout, ver.Info{
		Name:        "push",
		Description: "push a docker image to a remote docker repository",
		Version:     version,
		Commit:      commit,
		Date:        date,
	},
		os.Args[1:]...))
}
