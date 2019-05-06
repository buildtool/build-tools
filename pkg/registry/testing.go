// +build !prod

package registry

import (
  "context"
  "docker.io/go-docker/api/types"
  "docker.io/go-docker/api/types/registry"
)

type MockDocker struct {
  Username      string
  Password      string
  ServerAddress string
}

func (m *MockDocker) RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error) {
  m.Username = auth.Username
  m.Password = auth.Password
  m.ServerAddress = auth.ServerAddress
  return registry.AuthenticateOKBody{Status: "Logged in"}, nil
}

var _ DockerClient = &MockDocker{}
