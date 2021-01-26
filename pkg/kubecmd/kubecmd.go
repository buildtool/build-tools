package kubecmd

import (
	"flag"
	"fmt"
	"github.com/buildtool/build-tools/pkg/config"
	"io"
)

func Kubecmd(dir string, out io.Writer, args ...string) *string {
	var context, namespace string
	const (
		contextUsage   = "override the context for default deployment target"
		namespaceUsage = "override the namespace for default deployment target"
	)
	set := flag.NewFlagSet("deploy", flag.ExitOnError)
	set.Usage = func() {
		_, _ = fmt.Fprintf(out, "Usage: deploy [options] <target>\n\nFor example `deploy --context test-cluster --namespace test prod` would deploy to namespace `test` in the `test-cluster` but assuming to use the `prod` configuration files (if present)\n\nOptions:\n")
		set.PrintDefaults()
	}
	set.StringVar(&context, "context", "", contextUsage)
	set.StringVar(&context, "c", "", contextUsage+" (shorthand)")
	set.StringVar(&namespace, "namespace", "", namespaceUsage)
	set.StringVar(&namespace, "n", "", namespaceUsage+" (shorthand)")

	_ = set.Parse(args)

	if set.NArg() < 1 {
		set.Usage()
	} else {
		target := set.Args()[0]
		if cfg, err := config.Load(dir, out); err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
		} else {
			if env, err := cfg.CurrentTarget(target); err != nil {
				_, _ = fmt.Fprintln(out, err.Error())
			} else {
				if context != "" {
					env.Context = context
				}
				if namespace != "" {
					env.Namespace = namespace
				}

				cmd := fmt.Sprintf("kubectl --context %s --namespace %s", env.Context, env.Namespace)
				return &cmd
			}
		}
	}

	return nil
}
