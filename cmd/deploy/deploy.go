package main

import (
	"flag"
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"gitlab.com/sparetimecoders/build-tools/pkg/deploy"
	"gitlab.com/sparetimecoders/build-tools/pkg/kubectl"
	"os"
	"time"
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

		if cfg, err := config.Load(dir); err != nil {
			fmt.Println(err.Error())
		} else {
			if env, err := cfg.CurrentEnvironment(environment); err != nil {
				fmt.Println(err.Error())
			} else {
				if context != "" {
					env.Context = context
				}
				if namespace != "" {
					env.Namespace = namespace
				}
				ci := cfg.CurrentCI()
				tstamp := time.Now().Format(time.RFC3339)
				if err := deploy.Deploy(dir, ci.Commit(), tstamp, kubectl.New(env)); err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	}
}
