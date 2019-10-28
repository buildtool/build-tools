package deploy

import (
	"fmt"
	"github.com/sparetimecoders/build-tools/pkg/kubectl"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Deploy(dir, commit, buildName, timestamp, targetEnvironment string, client kubectl.Kubectl, out, eout io.Writer) error {
	deploymentFiles := filepath.Join(dir, "k8s")
	if err := processDir(deploymentFiles, commit, timestamp, targetEnvironment, client, out, eout); err != nil {
		return err
	}

	if client.DeploymentExists(buildName) {
		if !client.RolloutStatus(buildName) {
			_, _ = fmt.Fprint(out, "Rollout failed. Fetching events.")
			_, _ = fmt.Fprint(out, client.DeploymentEvents(buildName))
			_, _ = fmt.Fprint(out, client.PodEvents(buildName))
			return fmt.Errorf("failed to rollout")
		}
	}
	return nil
}

func processDir(dir, commit, timestamp, targetEnvironment string, client kubectl.Kubectl, out, eout io.Writer) error {
	if infos, err := ioutil.ReadDir(dir); err == nil {
		for _, info := range infos {
			if info.Name() == targetEnvironment && info.IsDir() {
				if err := processDir(filepath.Join(dir, info.Name()), commit, timestamp, targetEnvironment, client, out, eout); err != nil {
					return err
				}
			} else if fileIsForEnvironment(info, targetEnvironment) {
				if file, err := os.Open(filepath.Join(dir, info.Name())); err != nil {
					return err
				} else {
					if err := processFile(file, commit, timestamp, client); err != nil {
						return err
					}
				}
			} else if fileIsScriptForEnvironment(info, targetEnvironment) {
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

func fileIsForEnvironment(info os.FileInfo, env string) bool {
	return strings.HasSuffix(info.Name(), fmt.Sprintf("-%s.yaml", env)) || (strings.HasSuffix(info.Name(), ".yaml") && !strings.Contains(info.Name(), "-"))
}

func fileIsScriptForEnvironment(info os.FileInfo, env string) bool {
	return strings.HasSuffix(info.Name(), fmt.Sprintf("-%s.sh", env)) || (strings.HasSuffix(info.Name(), ".sh") && !strings.Contains(info.Name(), "-"))
}
