package registry

import "gitlab.com/sparetimecoders/build-tools/pkg/docker"

type Registry interface {
  identify() bool
  Login(client docker.Client) bool
  RegistryUrl() string
  Create() bool
  Validate() bool
}

var registries = []Registry{&dockerhub{}, &ecr{}, &gitlab{}, &quay{}}

func Identify() Registry {
  for _, reg := range registries {
    if reg.identify() {
      return reg
    }
  }
  return nil
}
