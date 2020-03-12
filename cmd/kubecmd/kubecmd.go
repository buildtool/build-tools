package main

import (
	"fmt"
	"github.com/buildtool/build-tools/pkg/kubecmd"
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
		if cmd := kubecmd.Kubecmd(dir, colorable.NewColorableStderr(), os.Args[1:]...); cmd != nil {
			_, _ = fmt.Fprintf(out, *cmd)
		}
	}
}
