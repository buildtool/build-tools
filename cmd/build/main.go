package main

import (
  "context"
  "docker.io/go-docker"
  "docker.io/go-docker/api/types"
  "flag"
  "fmt"
  "github.com/docker/docker/pkg/archive"
  "gitlab.com/sparetimecoders/build-tools/pkg/ci"
  "gitlab.com/sparetimecoders/build-tools/pkg/registry"
  "io/ioutil"
  "log"
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

  currentCI := ci.Identify()
  currentRegistry := registry.Identify()

  currentRegistry.Login(client)

  buildContext, err := archive.TarWithOptions(".", &archive.TarOptions{})
  if err != nil {
    panic(err)
  }

  tags := []string{
    tag(currentRegistry.RegistryUrl(), currentCI.Commit()),
    tag(currentRegistry.RegistryUrl(), currentCI.BranchReplaceSlash()),
  }
  if currentCI.Branch() == "master" {
    tags = append(tags, tag(currentRegistry.RegistryUrl(), "latest"))
  }
  response, err := client.ImageBuild(context.Background(), buildContext, types.ImageBuildOptions{
    Dockerfile: dockerfile,
    Memory:     3 * 1024 * 1024 * 1024,
    MemorySwap: -1,
    Remove:     true,
    ShmSize:    256 * 1024 * 1024,
    Tags:       tags,
  })

  if err != nil {
    panic(err)
  } else {
    bytes, _ := ioutil.ReadAll(response.Body)
    log.Printf("%+v\n", string(bytes))
  }
}

func tag(registry, tag string) string {
  return fmt.Sprintf("%s:%s", registry, tag)
}
