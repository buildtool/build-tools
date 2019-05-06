package registry

import (
  "context"
  "docker.io/go-docker"
  "docker.io/go-docker/api/types"
  "docker.io/go-docker/api/types/registry"
)

type Registry interface {
  identify() bool
  Login(client DockerClient) bool
  RegistryUrl() string
  Create() bool
  Validate() bool
}

type DockerClient interface {
  RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error)
}

var _ DockerClient = &docker.Client{}

var registries = []Registry{&dockerhub{}, &ecr{}, &gitlab{}, &quay{}}

func Identify() Registry {
  for _, reg := range registries {
    if reg.identify() {
      return reg
    }
  }
  return nil
}
