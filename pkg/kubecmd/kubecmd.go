package kubecmd

import (
	"fmt"
	"io"

	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/version"
)

type Args struct {
	args.Globals
	Target    string `arg:"" name:"target" help:"the target in the .buildtools.yaml"`
	Context   string `name:"context" short:"c" help:"override the context for default deployment target" default:""`
	Namespace string `name:"namespace" short:"n" help:"override the namespace for default deployment target" default:""`
}

func Kubecmd(dir string, out, eout io.Writer, info version.Info, osArgs ...string) *string {
	var kubeCmdArgs Args
	err := args.ParseArgs(dir, out, eout, osArgs, info, &kubeCmdArgs)
	if err != nil {
		return nil
	}
	if cfg, err := config.Load(dir, out); err != nil {
		_, _ = fmt.Fprintln(out, err.Error())
	} else {
		if env, err := cfg.CurrentTarget(kubeCmdArgs.Target); err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
		} else {
			if kubeCmdArgs.Context != "" {
				env.Context = kubeCmdArgs.Context
			}
			if kubeCmdArgs.Namespace != "" {
				env.Namespace = kubeCmdArgs.Namespace
			}

			if len(env.Namespace) == 0 {
				env.Namespace = "default"
			}

			cmd := fmt.Sprintf("kubectl --context %s --namespace %s", env.Context, env.Namespace)
			return &cmd
		}
	}

	return nil
}
