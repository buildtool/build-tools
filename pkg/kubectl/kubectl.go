package kubectl

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"io"
	"k8s.io/kubernetes/pkg/kubectl/cmd"
	"os"
)

type Kubectl interface {
	Apply(input io.Reader, args ...string) error
	Environment() *config.Environment
}

type kubectl struct {
	context     string
	namespace   string
	environment *config.Environment
}

func New(environment *config.Environment) Kubectl {
	ns := environment.Namespace
	if len(ns) == 0 {
		ns = "default"
	}
	return &kubectl{environment.Context, ns, environment}
}

func (kubectl) Apply(input io.Reader, args ...string) error {
	c := cmd.NewDefaultKubectlCommandWithArgs(nil, args, input, os.Stdout, os.Stderr)
	c.SetArgs(args)
	return c.Execute()
}

func (k kubectl) Environment() *config.Environment {
	return k.environment
}

var _ Kubectl = &kubectl{}
