// MIT License
//
// Copyright (c) 2021 buildtool
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package kubectl

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/cmd"
	"k8s.io/kubectl/pkg/cmd/plugin"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/buildtool/build-tools/pkg/cli"
	"github.com/buildtool/build-tools/pkg/config"
)

var kubectlVerbosityLevel = 6

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

func init() {
	// To enable flags for kubectl like --v
	klog.InitFlags(flag.CommandLine)
}

var newKubectlCmd = func(in io.Reader, out, errout io.Writer, args []string) *cobra.Command {
	c := cmd.NewKubectlCommand(cmd.KubectlOptions{
		PluginHandler: cmd.NewDefaultPluginHandler(plugin.ValidPluginFilenamePrefixes),
		ConfigFlags:   genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag(),
		IOStreams:     genericclioptions.IOStreams{In: in, Out: out, ErrOut: errout},
	})
	c.SetArgs(args)
	c.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	return c
}

func New(target *config.Target) Kubectl {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")

	args := argsFromTarget(target, name)

	out := cli.NewWriter(log.Log)
	return &kubectl{args: args, tempDir: name, out: out}
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
	err := os.WriteFile(kubeconfigFile, content, 0777)
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

	if cli.Verbose(log.Log) {
		args = append(args, fmt.Sprintf("--v=%d", kubectlVerbosityLevel))
	}
	return
}

func (k kubectl) Apply(input string) error {
	file := filepath.Join(k.tempDir, "content.yaml")
	err := os.WriteFile(file, []byte(input), 0777)
	if err != nil {
		return err
	}

	args := append(k.defaultArgs(), "apply", "-f", file)
	c := newKubectlCmd(os.Stdin, k.out, k.out, args)
	return c.Execute()
}

func (k kubectl) Cleanup() {
	_ = os.RemoveAll(k.tempDir)
}

func (k kubectl) DeploymentExists(name string) bool {
	args := k.defaultArgs()
	args = append(args, "get", "deployment", name, "--ignore-not-found")
	log.Debugf("kubectl %s\n", strings.Join(args, " "))
	buffer := bytes.Buffer{}
	c := newKubectlCmd(os.Stdin, &buffer, &buffer, args)
	_ = c.Execute()
	return buffer.Len() > 0
}

func (k kubectl) RolloutStatus(name, timeout string) bool {
	args := k.defaultArgs()
	args = append(args, "rollout", "status", "deployment", fmt.Sprintf("--timeout=%s", timeout), name)
	log.Debugf("kubectl %s\n", strings.Join(args, " "))
	c := newKubectlCmd(os.Stdin, k.out, k.out, args)
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
	log.Debugf("kubectl %s\n", strings.Join(args, " "))
	buffer := bytes.Buffer{}
	c := newKubectlCmd(os.Stdin, &buffer, &buffer, args)
	if err := c.Execute(); err != nil {
		return err.Error()
	}
	return k.extractEvents(buffer.String())
}

func (k kubectl) PodEvents(name string) string {
	args := k.defaultArgs()
	args = append(args, "describe", "pods", "-l", fmt.Sprintf("app=%s", name), "--show-events=true")
	log.Debugf("kubectl %s\n", strings.Join(args, " "))
	buffer := bytes.Buffer{}
	c := newKubectlCmd(os.Stdin, &buffer, &buffer, args)
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
