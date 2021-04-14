package deploy

import (
	"fmt"
	"time"

	"github.com/liamg/tml"

	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/ci"
	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/kubectl"
	"github.com/buildtool/build-tools/pkg/version"

	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Args struct {
	args.Globals
	Target    string `arg:"" name:"target" help:"the target in the .buildtools.yaml" default:"local"`
	Context   string `name:"context" short:"c" help:"override the context for default deployment target" default:""`
	Namespace string `name:"namespace" short:"n" help:"override the namespace for default deployment target" default:""`
	Tag       string `name:"tag" short:"n" help:"override the tag to deploy, not using the CI or VCS evaluated value" default:""`
	Timeout   string `name:"timeout" short:"t" help:"override the default deployment timeout (2 minutes). 0 means forever, all other values should contain a corresponding time unit (e.g. 1s, 2m, 3h)" default:"2m"`
}

func DoDeploy(dir string, out, eout io.Writer, info version.Info, osArgs ...string) int {
	var deployArgs Args
	err := args.ParseArgs(out, eout, osArgs, info, &deployArgs)
	if err != nil {
		if err != args.Done {
			return -1
		} else {
			return 0
		}
	}

	if cfg, err := config.Load(dir, out); err != nil {
		_, _ = fmt.Fprintln(out, err.Error())
		return -1
	} else {
		var env *config.Target
		if env, err = cfg.CurrentTarget(deployArgs.Target); err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
			env = &config.Target{}
		}
		if deployArgs.Context != "" {
			env.Context = deployArgs.Context
		}
		if env.Context == "" {
			_, _ = fmt.Fprintf(out, "context is mandatory, not found in configuration for %s and not passed as parameter\n", deployArgs.Target)
			return -5
		}
		if deployArgs.Namespace != "" {
			env.Namespace = deployArgs.Namespace
		}
		currentCI := cfg.CurrentCI()
		if deployArgs.Tag == "" {
			if !ci.IsValid(currentCI) {
				_, _ = fmt.Fprintln(out, tml.Sprintf("Commit and/or branch information is <red>missing</red>. Perhaps your not in a Git repository or forgot to set environment variables?"))
				return -3
			}
			deployArgs.Tag = currentCI.Commit()
		} else {
			_, _ = fmt.Fprintln(out, tml.Sprintf("Using passed tag <green>%s</green> to deploy", deployArgs.Tag))
		}

		tstamp := time.Now().Format(time.RFC3339)
		client := kubectl.New(env, out, eout)
		defer client.Cleanup()
		if err := Deploy(dir, currentCI.BuildName(), tstamp, client, out, eout, deployArgs); err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
			return -4

		}
	}
	return 0
}

func Deploy(dir, buildName, timestamp string, client kubectl.Kubectl, out, eout io.Writer, deployArgs Args) error {
	deploymentFiles := filepath.Join(dir, "k8s")
	if err := processDir(deploymentFiles, deployArgs.Tag, timestamp, deployArgs.Target, client, out, eout); err != nil {
		return err
	}

	if client.DeploymentExists(buildName) {
		if !client.RolloutStatus(buildName, deployArgs.Timeout) {
			_, _ = fmt.Fprint(out, "Rollout failed. Fetching events.")
			_, _ = fmt.Fprint(out, client.DeploymentEvents(buildName))
			_, _ = fmt.Fprint(out, client.PodEvents(buildName))
			return fmt.Errorf("failed to rollout")
		}
	}
	return nil
}

func processDir(dir, commit, timestamp, target string, client kubectl.Kubectl, out, eout io.Writer) error {
	if infos, err := ioutil.ReadDir(dir); err == nil {
		for _, info := range infos {
			if fileIsForTarget(info, target) {
				if file, err := os.Open(filepath.Join(dir, info.Name())); err != nil {
					return err
				} else {
					if err := processFile(file, commit, timestamp, client); err != nil {
						return err
					}
				}
			} else if fileIsScriptForTarget(info, target, dir) {
				if err := execFile(filepath.Join(dir, info.Name()), out, eout); err != nil {
					return err
				}
			}
		}
		return nil
	} else {
		return err
	}
}

func execFile(file string, out, eout io.Writer) error {
	cmd := exec.Command(file)
	cmd.Stdout = out
	cmd.Stderr = eout
	return cmd.Run()
}

func processFile(file *os.File, commit, timestamp string, client kubectl.Kubectl) error {
	if bytes, err := ioutil.ReadAll(file); err != nil {
		return err
	} else {
		content := string(bytes)
		r := strings.NewReplacer("${COMMIT}", commit, "${TIMESTAMP}", timestamp)
		if err := client.Apply(r.Replace(content)); err != nil {
			return err
		}
		return nil
	}
}

func fileIsForTarget(info os.FileInfo, env string) bool {
	return strings.HasSuffix(info.Name(), fmt.Sprintf("-%s.yaml", env)) || (strings.HasSuffix(info.Name(), ".yaml") && !strings.Contains(info.Name(), "-"))
}

func fileIsScriptForTarget(info os.FileInfo, env, dir string) bool {
	return strings.HasSuffix(info.Name(), fmt.Sprintf("-%s.sh", env))
}
