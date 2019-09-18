package push

import (
	docker2 "docker.io/go-docker"
	"flag"
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Push(dir string) int {
	var dockerfile string
	const (
		defaultDockerfile = "Dockerfile"
		usage             = "name of the Dockerfile to use"
	)
	set := flag.NewFlagSet("push", flag.ExitOnError)
	set.StringVar(&dockerfile, "file", defaultDockerfile, usage)
	set.StringVar(&dockerfile, "f", defaultDockerfile, usage+" (shorthand)")
	_ = set.Parse(os.Args)
	client, err := docker2.NewEnvClient()
	if err != nil {
		fmt.Println(err.Error())
		return -1
	}
	err = doPush(client, dir, dockerfile, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Println(err.Error())
		return -2
	}
	return 0
}

func doPush(client docker.Client, dir, dockerfile string, out, eout io.Writer) error {
	cfg, err := config.Load(dir, out)
	if err != nil {
		return err
	}
	currentCI, err := cfg.CurrentCI()
	if err != nil {
		return err
	}
	currentRegistry, err := cfg.CurrentRegistry()
	if err != nil {
		return err
	}

	if err := currentRegistry.Login(client, out); err != nil {
		return err
	}

	auth := currentRegistry.GetAuthInfo()

	if err := currentRegistry.Create(currentCI.BuildName()); err != nil {
		return err
	}

	content, err := ioutil.ReadFile(filepath.Join(dir, dockerfile))
	if err != nil {
		return err
	}
	stages := docker.FindStages(string(content))

	var tags []string
	for _, stage := range stages {
		tags = append(tags, docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), stage))
	}

	tags = append(tags,
		docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), currentCI.Commit()),
		docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), currentCI.BranchReplaceSlash()),
	)
	if currentCI.Branch() == "master" {
		tags = append(tags, docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), "latest"))
	}

	for _, tag := range tags {
		if err := currentRegistry.PushImage(client, auth, tag, out, eout); err != nil {
			return err
		}
	}
	return nil
}
