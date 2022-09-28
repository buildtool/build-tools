// MIT License
//
// Copyright (c) 2021 buildtool
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

package registry

import (
	"fmt"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg/docker"
)

func TestGithub_Name(t *testing.T) {
	registry := &Github{Repository: "repo", Username: "user", Password: "token"}

	assert.Equal(t, "Github", registry.Name())
}

func TestGithub_Configured(t *testing.T) {
	registry := &Github{Repository: "repo", Username: "user", Password: "token"}

	assert.True(t, registry.Configured())
}

func TestGithub_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &Github{Repository: "repo", Username: "user", Token: "token"}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.Nil(t, err)
	assert.Equal(t, "user", client.Username)
	assert.Equal(t, "token", client.Password)
	assert.Equal(t, "ghcr.io", client.ServerAddress)
	logMock.Check(t, []string{"debug: Logged in\n"})
}

func TestGithub_LoginError(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &Github{}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.EqualError(t, err, "invalid username/password")
	logMock.Check(t, []string{})
}

func TestGithub_GetAuthInfo(t *testing.T) {
	registry := &Github{Repository: "repo", Username: "user", Password: "token"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6InVzZXIiLCJwYXNzd29yZCI6InRva2VuIiwic2VydmVyYWRkcmVzcyI6ImdoY3IuaW8ifQ==", auth)
}

func TestGithub_Create(t *testing.T) {
	registry := &Github{}
	err := registry.Create("repo")
	assert.Nil(t, err)
}

func TestGithub_RegistryUrl(t *testing.T) {
	registry := &Github{Repository: "org/repo", Username: "user", Password: "token"}

	assert.Equal(t, "ghcr.io/org/repo", registry.RegistryUrl())
}
