package main

import (
  "flag"
  "gitlab.com/sparetimecoders/build-tools/pkg/ci"
  "log"
  "strings"
)

var dockerfile string

func init() {
  const (
    defaultDockerfile = "Dockerfile"
    usage = "name of the Dockerfile to use"
  )

  flag.StringVar(&dockerfile, "file", defaultDockerfile, usage)
  flag.StringVar(&dockerfile, "f", defaultDockerfile, usage + " (shorthand)")
}

func main() {
  flag.Parse()

  log.Println(dockerfile)
  log.Println(strings.Join(flag.Args(), " "))

  currentCI := ci.Identify()

  log.Printf("%+v\n", currentCI)
}
