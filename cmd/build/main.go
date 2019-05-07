package main

import (
  "docker.io/go-docker"
  "flag"
  "github.com/docker/docker/pkg/archive"
  "gitlab.com/sparetimecoders/build-tools/pkg/build"
)

var dockerfile string

func init() {
  const (
    defaultDockerfile = "Dockerfile"
    usage             = "name of the Dockerfile to use"
  )

  flag.StringVar(&dockerfile, "file", defaultDockerfile, usage)
  flag.StringVar(&dockerfile, "f", defaultDockerfile, usage+" (shorthand)")
}

func main() {
  flag.Parse()

  client, err := docker.NewEnvClient()
  if err != nil {
    panic(err)
  }

  buildContext, err := archive.TarWithOptions(".", &archive.TarOptions{})
  if err != nil {
    panic(err)
  }

  err = build.Build(client, buildContext, dockerfile)
}
