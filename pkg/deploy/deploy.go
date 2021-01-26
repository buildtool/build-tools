package deploy

import (
	"fmt"
	"github.com/buildtool/build-tools/pkg/kubectl"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Deploy(dir, commit, buildName, timestamp, target string, client kubectl.Kubectl, out, eout io.Writer, timeout string) error {
	deploymentFiles := filepath.Join(dir, "k8s")
	if err := processDir(deploymentFiles, commit, timestamp, target, client, out, eout); err != nil {
		return err
	}

	if client.DeploymentExists(buildName) {
		if !client.RolloutStatus(buildName, timeout) {
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
