package push

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/buildtool/build-tools/pkg/ci"
	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/docker"
	docker2 "github.com/docker/docker/client"
	"github.com/liamg/tml"
)

func Push(dir string, out, eout io.Writer, args ...string) int {
	var dockerfile string
	const (
		defaultDockerfile = "Dockerfile"
		usage             = "name of the Dockerfile to use"
	)
	set := flag.NewFlagSet("push", flag.ContinueOnError)
	set.StringVar(&dockerfile, "file", defaultDockerfile, usage)

	if err := set.Parse(args); err != nil {
		return -1
	}
	client, err := docker2.NewClientWithOpts(docker2.FromEnv)
	if err != nil {
		_, _ = fmt.Fprintln(eout, tml.Sprintf("<red>%s</red>", err.Error()))
		return -1
	}
	cfg, err := config.Load(dir, out)
	if err != nil {
		_, _ = fmt.Fprintln(eout, tml.Sprintf("<red>%s</red>", err.Error()))
		return -2
	}
	return doPush(client, cfg, dir, dockerfile, out, eout)
}

func doPush(client docker.Client, cfg *config.Config, dir, dockerfile string, out, eout io.Writer) int {
	currentCI := cfg.CurrentCI()
	currentRegistry := cfg.CurrentRegistry()

	if err := currentRegistry.Login(client, out); err != nil {
		_, _ = fmt.Fprintln(eout, tml.Sprintf("<red>%s</red>", err.Error()))
		return -3
	}

	auth := currentRegistry.GetAuthInfo()

	if err := currentRegistry.Create(currentCI.BuildName()); err != nil {
		_, _ = fmt.Fprintln(eout, tml.Sprintf("<red>%s</red>", err.Error()))
		return -4
	}

	content, err := ioutil.ReadFile(filepath.Join(dir, dockerfile))
	if err != nil {
		_, _ = fmt.Fprintln(eout, tml.Sprintf("<red>%s</red>", err.Error()))
		return -5
	}
	stages := docker.FindStages(string(content))

	var tags []string
	for _, stage := range stages {
		tags = append(tags, docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), stage, eout))
	}

	if dockerTag := os.Getenv("DOCKER_TAG"); len(dockerTag) > 0 {
		tag := docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), dockerTag, eout)
		tags = append(tags, tag)
		_, _ = fmt.Fprintf(out, "overriding docker tags with value from env DOCKER_TAG %s\n", dockerTag)
	} else {
		if !ci.IsValid(currentCI) {
			_, _ = fmt.Fprint(eout, tml.Sprintf("Commit and/or branch information is <red>missing</red>. Perhaps your not in a Git repository or forgot to set environment variables?"))
			return -6
		}
		tags = append(tags,
			docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), currentCI.Commit(), eout),
			docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), currentCI.BranchReplaceSlash(), eout),
		)
		if currentCI.Branch() == "master" || currentCI.Branch() == "main" {
			tags = append(tags, docker.Tag(currentRegistry.RegistryUrl(), currentCI.BuildName(), "latest", eout))
		}
	}
	for _, tag := range tags {
		_, _ = fmt.Fprintln(out, tml.Sprintf("Pushing tag '<green>%s</green>'", tag))
		if err := currentRegistry.PushImage(client, auth, tag, out, eout); err != nil {
			_, _ = fmt.Fprintln(eout, tml.Sprintf("<red>%s</red>", err.Error()))
			return -7
		}
	}
	return 0
}
