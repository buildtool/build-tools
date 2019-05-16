package kubectl

import (
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"io/ioutil"
	"k8s.io/kubernetes/pkg/kubectl/cmd"
	"os"
	"path/filepath"
)

type Kubectl interface {
	Apply(input string) error
	Environment() *config.Environment
	Cleanup()
}

type kubectl struct {
	context     string
	namespace   string
	environment *config.Environment
	tempDir     string
}

var newKubectlCmd = cmd.NewKubectlCommand

func New(environment *config.Environment) Kubectl {
	ns := environment.Namespace
	if len(ns) == 0 {
		ns = "default"
	}
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	return &kubectl{context: environment.Context, namespace: ns, environment: environment, tempDir: name}
}

func (k kubectl) Apply(input string) error {
	file := filepath.Join(k.tempDir, "content.yaml")
	err := ioutil.WriteFile(file, []byte(input), 0777)
	if err != nil {
		return err
	}
	args := []string{"apply", "--context", k.environment.Context, "--namespace", k.environment.Namespace, "-f", file}
	c := newKubectlCmd(os.Stdin, os.Stdout, os.Stderr)
	c.SetArgs(args)
	return c.Execute()
}

func (k kubectl) Environment() *config.Environment {
	return k.environment
}

func (k kubectl) Cleanup() {
	_ = os.RemoveAll(k.tempDir)
}

var _ Kubectl = &kubectl{}
