package kubectl

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"io"
	"io/ioutil"
	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"os"
	"path/filepath"
	"strings"
)

type Kubectl interface {
	Apply(input string) error
	Cleanup()
	DeploymentExists(name string) bool
	RolloutStatus(name string) bool
	DeploymentEvents(name string) string
	PodEvents(name string) string
}

type kubectl struct {
	context   string
	namespace string
	tempDir   string
	out       io.Writer
	eout      io.Writer
}

var newKubectlCmd = cmd.NewKubectlCommand

func New(environment *config.Environment, out, eout io.Writer) Kubectl {
	ns := environment.Namespace
	if len(ns) == 0 {
		ns = "default"
	}
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	return &kubectl{context: environment.Context, namespace: ns, tempDir: name, out: out}
}

func (k kubectl) defaultArgs() []string {
	return []string{
		"--context",
		k.context,
		"--namespace",
		k.namespace,
	}
}

func (k kubectl) Apply(input string) error {
	file := filepath.Join(k.tempDir, "content.yaml")
	err := ioutil.WriteFile(file, []byte(input), 0777)
	if err != nil {
		return err
	}

	args := k.defaultArgs()
	args = append(args, "apply", "-f", file)
	c := newKubectlCmd(os.Stdin, os.Stdout, os.Stderr)
	c.SetArgs(args)
	return c.Execute()
}

func (k kubectl) Cleanup() {
	_ = os.RemoveAll(k.tempDir)
}

func (k kubectl) DeploymentExists(name string) bool {
	args := k.defaultArgs()
	args = append(args, "get", "deployment", name)
	_, _ = fmt.Fprintf(k.out, "kubectl %s\n", strings.Join(args, " "))
	c := newKubectlCmd(os.Stdin, k.out, k.eout)
	c.SetArgs(args)
	return c.Execute() == nil
}

func (k kubectl) RolloutStatus(name string) bool {
	args := k.defaultArgs()
	args = append(args, "rollout", "status", "deployment", "--timeout=2m", name)
	_, _ = fmt.Fprintf(k.out, "kubectl %s\n", strings.Join(args, " "))
	c := newKubectlCmd(os.Stdin, os.Stdout, os.Stderr)
	c.SetArgs(args)
	success := true
	cmdutil.BehaviorOnFatal(func(str string, code int) {
		fmt.Println(str)
		success = false
	})
	if err := c.Execute(); err != nil {
		success = false
	}
	return success
}

func (k kubectl) DeploymentEvents(name string) string {
	args := k.defaultArgs()
	args = append(args, "describe", "deployment", name, "--show-events=true")
	_, _ = fmt.Fprintf(k.out, "kubectl %s\n", strings.Join(args, " "))
	buffer := bytes.Buffer{}
	c := newKubectlCmd(os.Stdin, &buffer, &buffer)
	c.SetArgs(args)
	if err := c.Execute(); err != nil {
		return err.Error()
	}
	return k.extractEvents(buffer.String())
}

func (k kubectl) PodEvents(name string) string {
	args := k.defaultArgs()
	args = append(args, "describe", "pods", "-l", fmt.Sprintf("app=%s", name), "--show-events=true")
	_, _ = fmt.Fprintf(k.out, "kubectl %s\n", strings.Join(args, " "))
	buffer := bytes.Buffer{}
	c := newKubectlCmd(os.Stdin, &buffer, &buffer)
	c.SetArgs(args)
	if err := c.Execute(); err != nil {
		return err.Error()
	}
	return k.extractEvents(buffer.String())
}

func (k kubectl) extractEvents(output string) string {
	scanner := bufio.NewScanner(strings.NewReader(output))
	var events strings.Builder
	found := false
	for scanner.Scan() {
		text := scanner.Text()
		if found || (strings.Index(text, "Events:") == 0 && strings.Index(text, "<none>") == -1) {
			found = true
			events.WriteString(text)
			events.WriteString("\n")
		}
	}

	return events.String()
}

var _ Kubectl = &kubectl{}
