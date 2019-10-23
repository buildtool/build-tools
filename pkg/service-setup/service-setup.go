package service_setup

import (
	"flag"
	"fmt"
	"github.com/liamg/tml"
	"github.com/sparetimecoders/build-tools/pkg/config"
	scaffold2 "github.com/sparetimecoders/build-tools/pkg/config/scaffold"
	"github.com/sparetimecoders/build-tools/pkg/stack"
	"io"
	"sort"
	"strings"
)

func Setup(dir string, out io.Writer, args ...string) int {
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
		return -1
	}

	name := set.Args()[0]
	currentStack, exists := stack.Stacks[selectedStack]
	if !exists {
		var stackNames []string
		for k := range stack.Stacks {
			stackNames = append(stackNames, k)
		}
		sort.Strings(stackNames)
		_, _ = fmt.Fprint(out, tml.Sprintf("<red>Provided stack does not exist yet. Available stacks are: </red><white><bold>(%s)</bold></white>\n", strings.Join(stackNames, ", ")))
		return -2
	}
	cfg, err := config.Load(dir, out)
	if err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -3
	}

	if err := cfg.Scaffold.ValidateConfig(); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -4
	}

	return scaffold(cfg.Scaffold, dir, name, currentStack, out)
}

func scaffold(cfg *scaffold2.Config, dir, name string, stack stack.Stack, out io.Writer) int {
	if err := cfg.Configure(); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -5
	}
	if err := cfg.Validate(name); err != nil {
		_, _ = fmt.Fprintln(out, tml.Sprintf("<red>%s</red>", err.Error()))
		return -6
	}
	return cfg.Scaffold(dir, name, stack, out)
}
