// +build !prod

package docker

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
)

type MockDocker struct {
	Username      string
	Password      string
	ServerAddress string
	BuildContext  []io.Reader
	BuildOptions  []types.ImageBuildOptions
	Images        []string
	LoginError    error
	BuildCount    int
	BuildError    []error
	PushError     error
	PushOutput    *string
	BrokenOutput  bool
	ResponseError error
}

func (m *MockDocker) ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	defer func() { m.BuildCount = m.BuildCount + 1 }()
	m.BuildContext = append(m.BuildContext, buildContext)
	m.BuildOptions = append(m.BuildOptions, options)

	if m.BrokenOutput {
		return types.ImageBuildResponse{Body: ioutil.NopCloser(strings.NewReader(`{"code":123,`))}, nil
	}
	if m.ResponseError != nil {
		return types.ImageBuildResponse{Body: ioutil.NopCloser(strings.NewReader(fmt.Sprintf(`{"errorDetail":{"code":123,"message":"%v"}}`, m.ResponseError)))}, nil
	}
	if len(m.BuildError) > m.BuildCount && m.BuildError[m.BuildCount] != nil {
		return types.ImageBuildResponse{Body: ioutil.NopCloser(strings.NewReader(fmt.Sprintf(`{"errorDetail":{"code":123,"message":"%v"}}`, m.BuildError)))}, m.BuildError[m.BuildCount]
	}
	return types.ImageBuildResponse{Body: ioutil.NopCloser(strings.NewReader(`{"stream":"Build successful"}`))}, nil
}

func (m *MockDocker) ImagePush(ctx context.Context, image string, options types.ImagePushOptions) (io.ReadCloser, error) {
	m.Images = append(m.Images, image)

	if m.PushError != nil {
		return ioutil.NopCloser(strings.NewReader("Push error")), m.PushError
	}

	return ioutil.NopCloser(strings.NewReader(*m.PushOutput)), nil
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
