// MIT License
//
// Copyright (c) 2018 buildtool
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

package deploy

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/apex/log"

	"github.com/buildtool/build-tools/pkg/file"

	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/ci"
	"github.com/buildtool/build-tools/pkg/cli"
	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/kubectl"
	"github.com/buildtool/build-tools/pkg/version"
)

type Args struct {
	args.Globals
	Target    string `arg:"" name:"target" help:"the target in the .buildtools.yaml"`
	Context   string `name:"context" short:"c" help:"override the context for default deployment target" default:""`
	Namespace string `name:"namespace" short:"n" help:"override the namespace for default deployment target" default:""`
	Tag       string `name:"tag" help:"override the tag to deploy, not using the CI or VCS evaluated value" default:""`
	Timeout   string `name:"timeout" short:"t" help:"override the default deployment timeout (2 minutes). 0 means forever, all other values should contain a corresponding time unit (e.g. 1s, 2m, 3h)" default:"2m"`
	NoWait    bool   `name:"no-wait" help:"don't wait for deployment to become ready"`
}

func DoDeploy(dir string, info version.Info, osArgs ...string) int {
	var deployArgs Args
	err := args.ParseArgs(dir, osArgs, info, &deployArgs)
	if err != nil {
		if err != args.ErrDone {
			return -1
		} else {
			return 0
		}
	}

	if cfg, err := config.Load(dir); err != nil {
		log.Error(err.Error())
		return -1
	} else {
		var env *config.Target
		if env, err = cfg.CurrentTarget(deployArgs.Target); err != nil {
			log.Warnf("%v\n", err)
			env = &config.Target{}
		}
		if deployArgs.Context != "" {
			env.Context = deployArgs.Context
		}
		if env.Context == "" {
			log.Errorf("context is mandatory, not found in configuration for %s and not passed as parameter\n", deployArgs.Target)
			return -5
		}
		if env.Context == "in-cluster" {
			log.Info("Using empty context for in-cluster deploy\n")
			env.Context = ""
		}
		if deployArgs.Namespace != "" {
			env.Namespace = deployArgs.Namespace
		}
		currentCI := cfg.CurrentCI()
		if deployArgs.Tag == "" {
			if !ci.IsValid(currentCI) {
				log.Errorf("Commit and/or branch information is <red>missing</red>. Perhaps your not in a Git repository or forgot to set environment variables?")
				return -3
			}
			deployArgs.Tag = currentCI.Commit()
		} else {
			log.Infof("Using passed tag <green>%s</green> to deploy", deployArgs.Tag)
		}

		tstamp := time.Now().Format(time.RFC3339)
		client := kubectl.New(env)
		defer client.Cleanup()
		if err := Deploy(dir, cfg.CurrentRegistry().RegistryUrl(), currentCI.BuildName(), tstamp, client, deployArgs); err != nil {
			log.Error(err.Error())
			return -4

		}
	}
	return 0
}

func Deploy(dir, registryUrl, buildName, timestamp string, client kubectl.Kubectl, deployArgs Args) error {
	imageName := fmt.Sprintf("%s/%s:%s", registryUrl, buildName, deployArgs.Tag)

	deploymentFiles := filepath.Join(dir, "k8s")
	if err := processDir(deploymentFiles, deployArgs.Tag, timestamp, deployArgs.Target, imageName, client); err != nil {
		return err
	}

	if deployArgs.NoWait {
		log.Info("Not waiting for deployment to succeed\n")
		return nil
	}

	if client.DeploymentExists(buildName) {
		if !client.RolloutStatus(buildName, deployArgs.Timeout) {
			log.Error("Rollout failed. Fetching events.\n")
			log.Error(client.DeploymentEvents(buildName))
			log.Error(client.PodEvents(buildName))
			return fmt.Errorf("failed to rollout")
		}
	}
	return nil
}

func processDir(dir, commit, timestamp, target, imageName string, client kubectl.Kubectl) error {
	files, err := file.FindFilesForTarget(dir, target)
	if err != nil {
		return err
	}
	scripts, err := file.FindScriptsForTarget(dir, target)
	if err != nil {
		return err
	}
	for _, info := range files {
		if f, err := os.Open(filepath.Join(dir, info.Name())); err != nil {
			return err
		} else {
			if err := processFile(f, commit, timestamp, imageName, client); err != nil {
				return err
			}
		}
	}
	for _, info := range scripts {
		if err := execFile(filepath.Join(dir, info.Name())); err != nil {
			return err
		}
	}
	return nil
}

func execFile(file string) error {
	cmd := exec.Command(file)
	cmd.Stdout = cli.NewWriter(log.Log)
	cmd.Stderr = cli.NewWriter(log.Log)
	return cmd.Run()
}

func processFile(file *os.File, commit, timestamp, image string, client kubectl.Kubectl) error {
	if bytes, err := io.ReadAll(file); err != nil {
		return err
	} else {
		content := string(bytes)
		if len(strings.TrimSpace(content)) == 0 {
			log.Debugf("ignoring empty file '<yellow>%s</yellow>'\n", filepath.Base(file.Name()))
			return nil
		}
		r := strings.NewReplacer("${COMMIT}", commit, "${TIMESTAMP}", timestamp, "${IMAGE}", image)
		kubeContent := r.Replace(content)
		log.Debugf("trying to apply: \n---\n%s\n---\n", kubeContent)
		if err := client.Apply(kubeContent); err != nil {
			return err
		}
		return nil
	}
}
