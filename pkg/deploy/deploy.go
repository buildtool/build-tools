package deploy

import (
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/kubectl"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Deploy(dir, commit, timestamp string, client kubectl.Kubectl) error {
	deploymentFiles := filepath.Join(dir, "deployment_files")
	return processDir(deploymentFiles, commit, timestamp, client)
}

func processDir(dir, commit, timestamp string, client kubectl.Kubectl) error {
	env := client.Environment()
	if infos, err := ioutil.ReadDir(dir); err == nil {
		for _, info := range infos {
			if info.Name() == env.Name && info.IsDir() {
				if err := processDir(filepath.Join(dir, info.Name()), commit, timestamp, client); err != nil {
					return err
				}
			} else if fileIsForEnvironment(info, env.Name) {
				if file, err := os.Open(filepath.Join(dir, info.Name())); err != nil {
					return err
				} else {
					if err := processFile(file, commit, timestamp, client); err != nil {
						return err
					}
				}
			}
		}
		return nil
	} else {
		return err
	}
}

func processFile(file *os.File, commit, timestamp string, client kubectl.Kubectl) error {
	if bytes, err := ioutil.ReadAll(file); err != nil {
		return err
	} else {
		env := client.Environment()
		content := string(bytes)
		content = strings.ReplaceAll(content, "${COMMIT}", commit)
		content = strings.ReplaceAll(content, "${TIMESTAMP}", timestamp)
		if err := client.Apply(strings.NewReader(content), "apply", "--context", env.Context, "--namespace", env.Namespace, "-f", "-"); err != nil {
			return err
		}
		return nil
	}
}

func fileIsForEnvironment(info os.FileInfo, env string) bool {
	return strings.HasSuffix(info.Name(), fmt.Sprintf("-%s.yaml", env)) || (strings.HasSuffix(info.Name(), ".yaml") && !strings.Contains(info.Name(), "-"))
}
