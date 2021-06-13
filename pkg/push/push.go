package push

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/apex/log"

	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/version"

	docker2 "github.com/docker/docker/client"

	"github.com/buildtool/build-tools/pkg/ci"
	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/docker"
)

type Args struct {
	args.Globals
	Dockerfile string `name:"file" short:"f" help:"name of the Dockerfile to use." default:"Dockerfile"`
}

func Push(dir string, info version.Info, osArgs ...string) int {
	var pushArgs Args
	err := args.ParseArgs(dir, osArgs, info, &pushArgs)
	if err != nil {
		if err != args.Done {
			return -1
		} else {
			return 0
		}
	}

	client, err := docker2.NewClientWithOpts(docker2.FromEnv)
	if err != nil {
		log.Error(fmt.Sprintf("<red>%s</red>", err.Error()))
		return -1
	}
	cfg, err := config.Load(dir)
	if err != nil {
		log.Error(fmt.Sprintf("<red>%s</red>", err.Error()))
		return -2
	}
	return doPush(client, cfg, dir, pushArgs.Dockerfile)
}

func doPush(client docker.Client, cfg *config.Config, dir, dockerfile string) int {
	currentCI := cfg.CurrentCI()
	currentRegistry := cfg.CurrentRegistry()

	if err := currentRegistry.Login(client); err != nil {
		log.Error(fmt.Sprintf("<red>%s</red>", err.Error()))
		return -3
	}

	auth := currentRegistry.GetAuthInfo()

	if err := currentRegistry.Create(currentCI.BuildName()); err != nil {
		log.Error(fmt.Sprintf("<red>%s</red>", err.Error()))
		return -4
	}

	content, err := ioutil.ReadFile(filepath.Join(dir, dockerfile))
	if err != nil {
		log.Error(fmt.Sprintf("<red>%s</red>", err.Error()))
		return -5
	}
	stages := docker.FindStages(string(content))

	var tags []string
	for _, stage := range stages {
		tags = append(tags, docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), stage))
	}

	if !ci.IsValid(currentCI) {
		log.Error("Commit and/or branch information is <red>missing</red>. Perhaps your not in a Git repository or forgot to set environment variables?")
		return -6
	}
	tags = append(tags,
		docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), currentCI.Commit()),
		docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), currentCI.BranchReplaceSlash()),
	)
	if currentCI.Branch() == "master" || currentCI.Branch() == "main" {
		tags = append(tags, docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), "latest"))
	}
	for _, tag := range tags {
		log.Info(fmt.Sprintf("Pushing tag '<green>%s</green>'\n", tag))
		if err := currentRegistry.PushImage(client, auth, tag); err != nil {
			log.Error(fmt.Sprintf("<red>%s</red>", err.Error()))
			return -7
		}
	}
	return 0
}
