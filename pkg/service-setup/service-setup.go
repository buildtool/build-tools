package service_setup

import (
	"flag"
	"fmt"
	"github.com/liamg/tml"
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"gitlab.com/sparetimecoders/build-tools/pkg/stack"
	"io"
	"strings"
)

func Setup(dir string, out io.Writer, exit func(code int), args ...string) {
	var selectedStack string
	const (
		stackUsage = "stack to scaffold"
	)
	set := flag.NewFlagSet("service-setup", flag.ExitOnError)
	set.Usage = func() {
		_, _ = fmt.Fprint(out, tml.Sprintf("Usage: service-setup [options] <name>\n\nFor example <blue>`service-setup --stack go gosvc`</blue> would create a new repository and scaffold it as a Go-project\n\nOptions:\n"))
		set.PrintDefaults()
	}
	set.StringVar(&selectedStack, "stack", "none", stackUsage)
	set.StringVar(&selectedStack, "s", "none", stackUsage+" (shorthand)")

	_ = set.Parse(args)

	if set.NArg() < 1 {
		set.Usage()
	} else {
		name := set.Args()[0]
		if currentStack, exists := stack.Stacks[selectedStack]; exists {
			if cfg, err := config.Load(dir, out); err != nil {
				_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
				exit(-1)
			} else {
				cfg.Scaffold(dir, name, currentStack, out, exit)
			}
		} else {
			var stackNames []string
			for k := range stack.Stacks {
				stackNames = append(stackNames, k)
			}
			_, _ = fmt.Fprint(out, tml.Sprintf("<red>Provided stack does not exist yet. Available stacks are: </red><white><bold>(%s)</bold></white>\n", strings.Join(stackNames, ", ")))
		}
	}
}
