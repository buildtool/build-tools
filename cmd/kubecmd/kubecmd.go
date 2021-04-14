package main

import (
	"fmt"
	"os"

	"github.com/mattn/go-colorable"

	"github.com/buildtool/build-tools/pkg/kubecmd"
	ver "github.com/buildtool/build-tools/pkg/version"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	out     = colorable.NewColorableStdout()
	eout    = colorable.NewColorableStderr()
)

func main() {
	dir, _ := os.Getwd()
	if cmd := kubecmd.Kubecmd(dir, eout, eout, ver.Info{
		Name:        "kubecmd",
		Description: "Generates a kubectl command, using the configuration from .buildtools.yaml if found",
		Version:     version,
		Commit:      commit,
		Date:        date,
	}, os.Args[1:]...); cmd != nil {
		_, _ = fmt.Fprintf(out, *cmd)
	}
}
