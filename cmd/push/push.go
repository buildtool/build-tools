package main

import (
	"docker.io/go-docker"
	"flag"
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/push"
	"os"
)

func main() {
	var dockerfile string

	const (
		defaultDockerfile = "Dockerfile"
		usage             = "name of the Dockerfile to use"
	)

	set := flag.NewFlagSet("push", flag.ExitOnError)
	set.StringVar(&dockerfile, "file", defaultDockerfile, usage)
	set.StringVar(&dockerfile, "f", defaultDockerfile, usage+" (shorthand)")

	_ = set.Parse(os.Args)

	client, err := docker.NewEnvClient()
	if err != nil {
		fmt.Println(err.Error())
	}

	err = push.Push(client, dockerfile, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Println(err.Error())
	}
}
