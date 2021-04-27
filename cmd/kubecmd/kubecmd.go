package main

import (
	"fmt"
	"io"
	"os"

	"github.com/apex/log"

	"github.com/buildtool/build-tools/pkg/cli"
	"github.com/buildtool/build-tools/pkg/kubecmd"
	ver "github.com/buildtool/build-tools/pkg/version"
)

var (
	version             = "dev"
	commit              = "none"
	date                = "unknown"
	out     io.Writer   = os.Stdout
	handler log.Handler = cli.New(os.Stderr)
)

func main() {
	log.SetHandler(handler)
	dir, _ := os.Getwd()
	if cmd := kubecmd.Kubecmd(dir, ver.Info{
		Name:        "kubecmd",
		Description: "Generates a kubectl command, using the configuration from .buildtools.yaml if found",
		Version:     version,
		Commit:      commit,
		Date:        date,
	}, os.Args[1:]...); cmd != nil {
		_, _ = fmt.Fprintf(out, *cmd)
	}
}
