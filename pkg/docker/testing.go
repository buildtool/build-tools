// +build !prod

package docker

import (
  "context"
  "docker.io/go-docker/api/types"
  "docker.io/go-docker/api/types/registry"
  "io"
  "io/ioutil"
  "strings"
)

type MockDocker struct {
  Username      string
  Password      string
  ServerAddress string
  BuildContext  io.Reader
  BuildOptions  types.ImageBuildOptions
  BuildError    error
}

func (m *MockDocker) ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
  m.BuildContext = buildContext
  m.BuildOptions = options

  if m.BuildError != nil {
    return types.ImageBuildResponse{Body: ioutil.NopCloser(strings.NewReader("Build error"))}, m.BuildError
  }
  return types.ImageBuildResponse{Body: ioutil.NopCloser(strings.NewReader("Build successful"))}, nil
}

func (m *MockDocker) RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error) {
  m.Username = auth.Username
  m.Password = auth.Password
  m.ServerAddress = auth.ServerAddress
  return registry.AuthenticateOKBody{Status: "Logged in"}, nil
}

var _ Client = &MockDocker{}
