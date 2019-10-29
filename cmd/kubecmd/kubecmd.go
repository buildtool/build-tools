package main

import (
	"fmt"
	"github.com/buildtool/build-tools/pkg/kubecmd"
	ver "github.com/buildtool/build-tools/pkg/version"
	"io"
	"os"
)

var (
	version            = "dev"
	commit             = "none"
	date               = "unknown"
	exitFunc           = os.Exit
	out      io.Writer = os.Stdout
)

func main() {
	if ver.PrintVersionOnly(version, commit, date, out) {
		exitFunc(0)
	} else {
		dir, _ := os.Getwd()
		if cmd := kubecmd.Kubecmd(dir, os.Stderr, os.Args[1:]...); cmd != nil {
			fmt.Fprintf(out, *cmd)
		}
	}
}
