package kubectl

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/apex/log"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/kubectl/pkg/cmd"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/buildtool/build-tools/pkg/cli"
	"github.com/buildtool/build-tools/pkg/config"
)

type Kubectl interface {
	Apply(input string) error
	Cleanup()
	DeploymentExists(name string) bool
	RolloutStatus(name, timeout string) bool
	DeploymentEvents(name string) string
	PodEvents(name string) string
}

type kubectl struct {
	args    map[string]string
	tempDir string
	out     io.Writer
}

var newKubectlCmd = cmd.NewKubectlCommand

func New(target *config.Target) Kubectl {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")

	arg := argsFromTarget(target, name)

	out := cli.NewWriter(log.Log)
	return &kubectl{args: arg, tempDir: name, out: out}
}

func argsFromTarget(e *config.Target, tempDir string) map[string]string {
	kubeconfigArg := "kubeconfig"
	args := make(map[string]string)
	if len(e.Context) > 0 {
		args["context"] = e.Context
	}
	if len(e.Namespace) > 0 {
		args["namespace"] = e.Namespace
	}

	if file, exists, err := getKubeconfigFileFromEnvs(tempDir); err != nil {
		log.Errorf("%s", err.Error())
	} else if exists {
		args[kubeconfigArg] = file
	}
	if len(e.Kubeconfig) > 0 {
		args[kubeconfigArg] = e.Kubeconfig
	}
	if _, exists := args[kubeconfigArg]; exists {
		log.Debugf("Using kubeconfig: <green>'%s'</green>", args[kubeconfigArg])
	}

	return args
}

func writeKubeconfigFile(tempDir string, content []byte) (string, error) {
	kubeconfigFile := filepath.Join(tempDir, "kubeconfig")
	err := ioutil.WriteFile(kubeconfigFile, content, 0777)
	return kubeconfigFile, err
}

func (k kubectl) defaultArgs() (args []string) {
	var keys []string
	for key := range k.args {
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
	c := newKubectlCmd(os.Stdin, k.out, k.out)
	c.SetArgs(args)
	return c.Execute()
}

func (k kubectl) Cleanup() {
	_ = os.RemoveAll(k.tempDir)
}

func (k kubectl) DeploymentExists(name string) bool {
	args := k.defaultArgs()
	args = append(args, "get", "deployment", name, "--ignore-not-found")
	_, _ = fmt.Fprintf(k.out, "kubectl %s\n", strings.Join(args, " "))
	buffer := bytes.Buffer{}
	c := newKubectlCmd(os.Stdin, &buffer, &buffer)
	c.SetArgs(args)
	_ = c.Execute()
	return buffer.Len() > 0
}

func (k kubectl) RolloutStatus(name, timeout string) bool {
	args := k.defaultArgs()
	args = append(args, "rollout", "status", "deployment", fmt.Sprintf("--timeout=%s", timeout), name)
	_, _ = fmt.Fprintf(k.out, "kubectl %s\n", strings.Join(args, " "))
	c := newKubectlCmd(os.Stdin, k.out, k.out)
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
		if found || strings.Index(text, "Events:") == 0 && !strings.Contains(text, "<none>") {
			found = true
			events.WriteString(text)
			events.WriteString("\n")
		}
	}

	return events.String()
}

var _ Kubectl = &kubectl{}

const envKubeconfigContent = "KUBECONFIG_CONTENT"

func getKubeconfigFileFromEnvs(tempDir string) (string, bool, error) {
	if content, exists := getEnv(envKubeconfigContent); exists {
		log.Debugf("Parsing kubeconfig from env: %s\n", envKubeconfigContent)
		if decoded, err := base64.StdEncoding.DecodeString(content); err != nil {
			log.Debugf("Failed to decode BASE64, falling back to plaintext\n")
			file, err := writeKubeconfigFile(tempDir, []byte(content))
			return file, true, err
		} else {
			file, err := writeKubeconfigFile(tempDir, decoded)
			return file, true, err
		}
	}
	return "", false, nil
}

func getEnv(env string) (string, bool) {
	if value, exists := os.LookupEnv(env); exists {
		if len(value) > 0 {
			log.Debugf("Found content in env: %s\n", env)
			return value, true
		} else {
			log.Debugf("environment variable %s is set but has no value\n", env)
		}
	}
	return "", false
}
