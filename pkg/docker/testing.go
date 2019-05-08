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
	Images        []string
	LoginError    error
	BuildError    error
	PushError     error
}

func (m *MockDocker) ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	m.BuildContext = buildContext
	m.BuildOptions = options

	if m.BuildError != nil {
		return types.ImageBuildResponse{Body: ioutil.NopCloser(strings.NewReader("Build error"))}, m.BuildError
	}
	return types.ImageBuildResponse{Body: ioutil.NopCloser(strings.NewReader("Build successful"))}, nil
}

func (m *MockDocker) ImagePush(ctx context.Context, image string, options types.ImagePushOptions) (io.ReadCloser, error) {
	m.Images = append(m.Images, image)

	if m.PushError != nil {
		return ioutil.NopCloser(strings.NewReader("Push error")), m.PushError
	}

	return ioutil.NopCloser(strings.NewReader("Push successful")), nil
}

func (m *MockDocker) RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error) {
	m.Username = auth.Username
	m.Password = auth.Password
	m.ServerAddress = auth.ServerAddress
	if m.LoginError != nil {
		return registry.AuthenticateOKBody{Status: "Invalid username/password"}, m.LoginError
	}
	return registry.AuthenticateOKBody{Status: "Logged in"}, nil
}

var _ Client = &MockDocker{}
