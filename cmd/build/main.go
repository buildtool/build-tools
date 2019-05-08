package main

import (
	"docker.io/go-docker"
	"flag"
	"fmt"
	"github.com/docker/docker/pkg/archive"
	"gitlab.com/sparetimecoders/build-tools/pkg/build"
	"os"
)

func main() {
	var dockerfile string
	const (
		defaultDockerfile = "Dockerfile"
		usage             = "name of the Dockerfile to use"
	)

	set := flag.NewFlagSet("build", flag.ExitOnError)
	set.StringVar(&dockerfile, "file", defaultDockerfile, usage)
	set.StringVar(&dockerfile, "f", defaultDockerfile, usage+" (shorthand)")
	_ = set.Parse(os.Args)

	client, err := docker.NewEnvClient()
	if err != nil {
		fmt.Println(err.Error())
	}

	buildContext, err := archive.TarWithOptions(".", &archive.TarOptions{})
	if err != nil {
		fmt.Println(err.Error())
	}

	err = build.Build(client, buildContext, dockerfile)
	if err != nil {
		fmt.Println(err.Error())
	}
}
