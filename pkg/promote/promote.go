package promote

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"

	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/ci"
	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/version"
)

type Args struct {
	args.Globals
	Target     string `arg:"" name:"target" help:"the target in the .buildtools.yaml" default:""`
	Namespace  string `name:"namespace" short:"n" help:"override the namespace for default deployment target" default:""`
	Tag        string `name:"tag" help:"override the tag to deploy, not using the CI or VCS evaluated value" default:""`
	URL        string `name:"url" help:"override the URL to the Git repository where files will be generated" default:""`
	Path       string `name:"path" help:"override the path in the Git repository where files will be generated" default:""`
	User       string `name:"user" help:"username for Git access" default:"git"`
	PrivateKey string `name:"key" help:"private key for Git access (defaults to ~/.ssh/id_rsa)" default:""`
	Password   string `name:"password" help:"password for private key" default:""`
	Out        string `name:"out" short:"o" help:"write output to specified file instead of committing and pushing to Git" default:""`
}

func DoPromote(dir string, info version.Info, osArgs ...string) int {
	var promoteArgs Args
	err := args.ParseArgs(dir, osArgs, info, &promoteArgs)
	if err != nil {
		if err != args.Done {
			return -1
		} else {
			return 0
		}
	}

	if cfg, err := config.Load(dir); err != nil {
		log.Error(err.Error())
		return -1
	} else {
		var target *config.Gitops
		if target, err = cfg.CurrentGitops(promoteArgs.Target); err != nil {
			log.Error(err.Error())
			return -2
		}
		if promoteArgs.URL != "" {
			target.URL = promoteArgs.URL
		}
		if promoteArgs.Path != "" {
			target.Path = promoteArgs.Path
		}
		currentCI := cfg.CurrentCI()
		if promoteArgs.Tag == "" {
			if !ci.IsValid(currentCI) {
				log.Errorf("Commit and/or branch information is <red>missing</red>. Perhaps your not in a Git repository or forgot to set environment variables?")
				return -3
			}
			promoteArgs.Tag = currentCI.Commit()
		} else {
			log.Infof("Using passed tag <green>%s</green> to promote", promoteArgs.Tag)
		}

		tstamp := time.Now().Format(time.RFC3339)
		if err := Promote(dir, currentCI.BuildName(), tstamp, target, promoteArgs, cfg.Git); err != nil {
			log.Error(err.Error())
			return -4
		}
	}
	return 0
}

func Promote(dir, name, timestamp string, target *config.Gitops, args Args, gitConfig config.Git) error {
	deploymentFiles := filepath.Join(dir, "k8s")
	if _, err := os.Lstat(deploymentFiles); os.IsNotExist(err) {
		return fmt.Errorf("no deployment descriptors found in k8s directory")
	}

	log.Info("generating...")
	buffer := &bytes.Buffer{}
	if err := processDir(buffer, deploymentFiles, args.Tag, timestamp, args.Target); err != nil {
		return err
	}

	if args.Out == "" {
		privKey := "~/.ssh/id_rsa"
		if args.PrivateKey != "" {
			privKey = args.PrivateKey
		} else if gitConfig.Key != "" {
			privKey = gitConfig.Key
		}
		if strings.HasPrefix(privKey, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			privKey = fmt.Sprintf("%s%s", home, strings.TrimPrefix(privKey, "~"))
		}
		keys, err := ssh.NewPublicKeysFromFile(args.User, privKey, args.Password)
		if err != nil {
			return err
		}
		cloneDir, err := ioutil.TempDir(os.TempDir(), "build-tools")
		if err != nil {
			return err
		}
		defer func(path string) {
			_ = os.RemoveAll(path)
		}(cloneDir)
		repo, err := git.PlainClone(cloneDir, false, &git.CloneOptions{
			URL:  target.URL,
			Auth: keys,
		})
		if err != nil {
			return err
		}
		worktree, err := repo.Worktree()
		if err != nil {
			return err
		}

		err = os.MkdirAll(filepath.Join(cloneDir, target.Path, name), 0777)
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(cloneDir, target.Path, name, "deploy.yaml"), buffer.Bytes(), 0666)
		if err != nil {
			return err
		}
		_, err = worktree.Add(filepath.Join(target.Path, name, "deploy.yaml"))
		if err != nil {
			return err
		}

		hash, err := worktree.Commit(
			fmt.Sprintf("ci: promoting %s commit %s to %s", name, args.Tag, args.Target),
			&git.CommitOptions{
				Author: &object.Signature{
					Name:  ifEmpty(gitConfig.Name, "Buildtools"),
					Email: ifEmpty(gitConfig.Email, "git@buildtools.io"),
					When:  time.Now(),
				},
			},
		)
		if err != nil {
			return err
		}
		commit, err := repo.CommitObject(hash)
		if err != nil {
			return err
		}
		log.Infof("pushing commit %s to %s/%s/%s", commit.Hash, target.URL, target.Path, name)
		err = repo.Push(&git.PushOptions{
			Auth: keys,
		})
		if err != nil {
			return err
		}
	} else {
		err := os.WriteFile(args.Out, buffer.Bytes(), 0666)
		if err != nil {
			return err
		}
	}

	return nil
}

func ifEmpty(s string, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

func processDir(writer io.StringWriter, dir, commit, timestamp, target string) error {
	if infos, err := ioutil.ReadDir(dir); err == nil {
		for _, info := range infos {
			if fileIsForTarget(info, target) {
				log.Debugf("using file '<green>%s</green>' for target: <green>%s</green>\n", info.Name(), target)
				if file, err := os.Open(filepath.Join(dir, info.Name())); err != nil {
					return err
				} else {
					if err := processFile(writer, file, commit, timestamp); err != nil {
						return err
					}
				}
			} else {
				log.Debugf("not using file '<red>%s</red>' for target: <green>%s</green>\n", info.Name(), target)
			}
		}
		return nil
	} else {
		return err
	}
}

func processFile(writer io.StringWriter, file *os.File, commit, timestamp string) error {
	if buff, err := ioutil.ReadAll(file); err != nil {
		return err
	} else {
		content := string(buff)
		r := strings.NewReplacer("${COMMIT}", commit, "${TIMESTAMP}", timestamp)
		kubeContent := r.Replace(content)
		_, err := writer.WriteString(kubeContent)
		if err != nil {
			return err
		}
		_, err = writer.WriteString("\n---\n")
		if err != nil {
			return err
		}

		return nil
	}
}

func fileIsForTarget(info os.FileInfo, env string) bool {
	log.Debugf("considering file '<yellow>%s</yellow>' for target: <green>%s</green>\n", info.Name(), env)
	return strings.HasSuffix(info.Name(), fmt.Sprintf("-%s.yaml", env)) || (strings.HasSuffix(info.Name(), ".yaml") && !strings.Contains(info.Name(), "-"))
}
