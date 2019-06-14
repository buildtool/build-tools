package main

import (
	"docker.io/go-docker"
	"flag"
	"fmt"
	"github.com/docker/docker/pkg/archive"
	"gitlab.com/sparetimecoders/build-tools/pkg/build"
	d2 "gitlab.com/sparetimecoders/build-tools/pkg/docker"
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

	if client, err := docker.NewEnvClient(); err != nil {
		fmt.Println(err.Error())
	} else {
		if ignored, err := d2.ParseDockerignore(); err != nil {
			fmt.Println(err.Error())
		} else {
			if buildContext, err := archive.TarWithOptions(".", &archive.TarOptions{ExcludePatterns: ignored}); err != nil {
				fmt.Println(err.Error())
			} else {
				err = build.Build(client, buildContext, dockerfile, os.Stdout, os.Stderr)
				if err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	}
}
