// MIT License
//
// Copyright (c) 2018 buildtool
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

//go:build !prod

package docker

import (
	"context"
	"fmt"
	"io"
	"iter"
	"net"
	"strings"

	"github.com/moby/moby/api/types/jsonstream"
	mobyclient "github.com/moby/moby/client"
)

type MockDocker struct {
	Username      string
	Password      string
	ServerAddress string
	BuildContext  []io.Reader
	BuildOptions  []mobyclient.ImageBuildOptions
	Images        []string
	LoginError    error
	BuildCount    int
	BuildError    []error
	PushError     error
	PushOutput    *string
	BrokenOutput  bool
	ResponseError error
	ResponseBody  io.Reader
}

func (m *MockDocker) ImageBuild(_ context.Context, buildContext io.Reader, options mobyclient.ImageBuildOptions) (mobyclient.ImageBuildResult, error) {
	defer func() { m.BuildCount = m.BuildCount + 1 }()
	m.BuildContext = append(m.BuildContext, buildContext)
	m.BuildOptions = append(m.BuildOptions, options)

	if m.BrokenOutput {
		return mobyclient.ImageBuildResult{Body: io.NopCloser(strings.NewReader(`{"errorDetail":{"code":0,"message":"some message"}}`))}, nil
	}
	if m.ResponseError != nil {
		return mobyclient.ImageBuildResult{Body: io.NopCloser(strings.NewReader(fmt.Sprintf(`{"errorDetail":{"code":123,"message":"%v"}}`, m.ResponseError)))}, nil
	}
	if len(m.BuildError) > m.BuildCount && m.BuildError[m.BuildCount] != nil {
		return mobyclient.ImageBuildResult{Body: io.NopCloser(strings.NewReader(fmt.Sprintf(`{"errorDetail":{"code":123,"message":"%v"}}`, m.BuildError)))}, m.BuildError[m.BuildCount]
	}
	var body io.Reader = strings.NewReader(`{"stream":"Build successful"}`)
	if m.ResponseBody != nil {
		body = m.ResponseBody
	}
	return mobyclient.ImageBuildResult{Body: io.NopCloser(body)}, nil
}

func (m *MockDocker) ImagePush(_ context.Context, image string, _ mobyclient.ImagePushOptions) (mobyclient.ImagePushResponse, error) {
	m.Images = append(m.Images, image)

	if m.PushError != nil {
		return &mockPushResponse{ReadCloser: io.NopCloser(strings.NewReader("Push error"))}, m.PushError
	}

	return &mockPushResponse{ReadCloser: io.NopCloser(strings.NewReader(*m.PushOutput))}, nil
}

func (m *MockDocker) RegistryLogin(_ context.Context, options mobyclient.RegistryLoginOptions) (mobyclient.RegistryLoginResult, error) {
	m.Username = options.Username
	m.Password = options.Password
	m.ServerAddress = options.ServerAddress
	if m.LoginError != nil {
		return mobyclient.RegistryLoginResult{}, m.LoginError
	}
	return mobyclient.RegistryLoginResult{}, nil
}

func (m *MockDocker) DialHijack(context.Context, string, string, map[string][]string) (net.Conn, error) {
	return nil, nil
}

func (m *MockDocker) BuildCancel(context.Context, string, mobyclient.BuildCancelOptions) (mobyclient.BuildCancelResult, error) {
	return mobyclient.BuildCancelResult{}, nil
}

// mockPushResponse implements mobyclient.ImagePushResponse for testing.
type mockPushResponse struct {
	io.ReadCloser
}

func (m *mockPushResponse) JSONMessages(_ context.Context) iter.Seq2[jsonstream.Message, error] {
	return nil
}

func (m *mockPushResponse) Wait(_ context.Context) error {
	return nil
}

var _ Client = &MockDocker{}
