package version

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
)

func PrintVersionOnly(version, commit, date string, out io.Writer) bool {
	var printVersion bool
	set := flag.NewFlagSet("versions", flag.ContinueOnError)
	set.Usage = func() {}
	set.SetOutput(&bytes.Buffer{})

	set.BoolVar(&printVersion, "version", false, "if true, print version and exit")
	_ = set.Parse(os.Args[1:])
	if printVersion {
		_, _ = fmt.Fprintf(out, "Version: %v, commit %v, built at %v\n", version, commit, date)
	}
	return printVersion
}
