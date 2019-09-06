package main

import (
	"flag"
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"os"
)

func main() {
	var context, namespace string
	const (
		contextUsage   = "override the context for default environment deployment target"
		namespaceUsage = "override the namespace for default environment deployment target"
	)
	set := flag.NewFlagSet("deploy", flag.ExitOnError)
	set.Usage = func() {
		fmt.Printf("Usage: deploy [options] <environment>\n\nFor example `deploy --context test-cluster --namespace test prod` would deploy to namsepace `test` in the `test-cluster` but assuming to use the `prod` configuration files (if present)\n\nOptions:\n")
		set.PrintDefaults()
	}
	set.StringVar(&context, "context", "", contextUsage)
	set.StringVar(&context, "c", "", contextUsage+" (shorthand)")
	set.StringVar(&namespace, "namespace", "", namespaceUsage)
	set.StringVar(&namespace, "n", "", namespaceUsage+" (shorthand)")
	_ = set.Parse(os.Args[1:])

	if set.NArg() < 1 {
		set.Usage()
	} else {
		environment := set.Args()[0]
		dir, _ := os.Getwd()

		if cfg, err := config.Load(dir, os.Stderr); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
		} else {
			if env, err := cfg.CurrentEnvironment(environment); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err.Error())
			} else {
				if context != "" {
					env.Context = context
				}
				if namespace != "" {
					env.Namespace = namespace
				}

				fmt.Printf("kubectl --context %s --namespace %s\n", env.Context, env.Namespace)
			}
		}
	}
}
