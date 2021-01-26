package main

import (
	"flag"
	"fmt"
	"github.com/buildtool/build-tools/pkg/ci"
	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/deploy"
	"github.com/buildtool/build-tools/pkg/kubectl"
	ver "github.com/buildtool/build-tools/pkg/version"
	"github.com/liamg/tml"
	"github.com/mattn/go-colorable"
	"os"
	"time"
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
		exitFunc(doDeploy())
	}
}

func doDeploy() int {
	var context, namespace, timeout, tag string
	const (
		contextUsage   = "override the context for default target deployment target"
		tagUsage       = "override the tag to deploy, not using the CI or VCS evaluated value"
		namespaceUsage = "override the namespace for default deployment target"
		timeoutUsage   = "override the default deployment timeout (2 minutes). 0 means forever, all other values should contain a corresponding time unit (e.g. 1s, 2m, 3h)"
	)
	set := flag.NewFlagSet("deploy", flag.ContinueOnError)
	set.Usage = func() {
		_, _ = fmt.Fprintf(out, "Usage: deploy [options] <target>\n\nFor example `deploy --context test-cluster --namespace test prod` would deploy to namespace `test` in the `test-cluster` but assuming to use the `prod` configuration files (if present)\n\nOptions:\n")
		set.SetOutput(out)
		set.PrintDefaults()
	}
	set.StringVar(&context, "context", "", contextUsage)
	set.StringVar(&namespace, "namespace", "", namespaceUsage)
	set.StringVar(&timeout, "timeout", "", timeoutUsage)
	set.StringVar(&tag, "tag", "", tagUsage)
	if err := set.Parse(os.Args[1:]); err != nil {
		return -1
	}

	var target string
	if set.NArg() < 1 {
		target = "local"
	} else {
		target = set.Args()[0]
	}
	dir, _ := os.Getwd()

	if cfg, err := config.Load(dir, out); err != nil {
		_, _ = fmt.Fprintln(out, err.Error())
		return -1
	} else {
		var env *config.Target
		if env, err = cfg.CurrentTarget(target); err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
			env = &config.Target{}
		}
		if context != "" {
			env.Context = context
		}
		if env.Context == "" {
			_, _ = fmt.Fprintf(out, "context is mandatory, not found in configuration for %s and not passed as parameter\n", target)
			return -5
		}
		if namespace != "" {
			env.Namespace = namespace
		}
		if timeout == "" {
			timeout = "2m"
		}
		currentCI := cfg.CurrentCI()
		if tag == "" {
			if !ci.IsValid(currentCI) {
				_, _ = fmt.Fprintln(out, tml.Sprintf("Commit and/or branch information is <red>missing</red>. Perhaps your not in a Git repository or forgot to set environment variables?"))
				return -3
			}
			tag = currentCI.Commit()
		} else {
			_, _ = fmt.Fprintln(out, tml.Sprintf("Using passed tag <green>%s</green> to deploy", tag))
		}

		tstamp := time.Now().Format(time.RFC3339)
		client := kubectl.New(env, out, colorable.NewColorableStderr())
		defer client.Cleanup()
		if err := deploy.Deploy(dir, tag, currentCI.BuildName(), tstamp, target, client, out, colorable.NewColorableStderr(), timeout); err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
			return -4

		}
	}
	return 0
}
