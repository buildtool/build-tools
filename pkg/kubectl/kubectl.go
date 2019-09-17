package kubectl

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/liamg/tml"
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"io"
	"io/ioutil"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"os"
	"path/filepath"
	"sort"
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
	args    map[string]string
	tempDir string
	out     io.Writer
	eout    io.Writer
}

var newKubectlCmd = cmd.NewKubectlCommand

func New(environment *config.Environment, out, eout io.Writer) Kubectl {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")

	arg := argsFromEnvironment(environment, name, out)
	return &kubectl{args: arg, tempDir: name, out: out, eout: eout}
}

func argsFromEnvironment(e *config.Environment, tempDir string, out io.Writer) map[string]string {
	kubeConfigArg := "kubeconfig"
	args := make(map[string]string)
	if len(e.Context) > 0 {
		args["context"] = e.Context
	}
	if len(e.Namespace) > 0 {
		args["namespace"] = e.Namespace
	}
	if content, exist := os.LookupEnv(envKubeConfigContent); exist {
		// Not a file, create file from content
		kubeconfigFile := filepath.Join(tempDir, kubeConfigArg)
		_ = ioutil.WriteFile(kubeconfigFile, []byte(content), 0777)
		args[kubeConfigArg] = kubeconfigFile
	} else if len(e.Kubeconfig) > 0 {
		args[kubeConfigArg] = e.Kubeconfig
	}
	if _, exists := args[kubeConfigArg]; exists {
		_, _ = fmt.Fprintln(out, tml.Sprintf("Using kubeconfig: <green>'%s'</green>", args[kubeConfigArg]))
	}

	return args
}

func (k kubectl) defaultArgs() (args []string) {
	var keys []string
	for key, _ := range k.args {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		args = append(args, fmt.Sprintf("--%s", key), k.args[key])
	}

	return
}

func (k kubectl) Apply(input string) error {
	file := filepath.Join(k.tempDir, "content.yaml")
	err := ioutil.WriteFile(file, []byte(input), 0777)
	if err != nil {
		return err
	}

	args := append(k.defaultArgs(), "apply", "-f", file)
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

const envKubeConfigContent = "KUBECONFIG_CONTENT"
