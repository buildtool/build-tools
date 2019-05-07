package docker

import (
  "context"
  "docker.io/go-docker"
  "docker.io/go-docker/api/types"
  "docker.io/go-docker/api/types/registry"
  "io"
)

type Client interface {
  RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error)
  ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
}

var _ Client = &docker.Client{}
