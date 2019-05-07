package main

import (
	"docker.io/go-docker"
	"flag"
	"gitlab.com/sparetimecoders/build-tools/pkg/push"
)

func main() {
	var dockerfile string

	const (
		defaultDockerfile = "Dockerfile"
		usage             = "name of the Dockerfile to use"
	)

	flag.StringVar(&dockerfile, "file", defaultDockerfile, usage)
	flag.StringVar(&dockerfile, "f", defaultDockerfile, usage+" (shorthand)")

	flag.Parse()

	client, err := docker.NewEnvClient()
	if err != nil {
		panic(err)
	}

	err = push.Push(client, dockerfile)
	if err != nil {
		panic(err)
	}
}
