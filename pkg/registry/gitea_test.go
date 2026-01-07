// MIT License
//
// Copyright (c) 2026 buildtool
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

func TestGitea_Name(t *testing.T) {
	registry := &Gitea{Registry: "gitea.example.com", Repository: "org/repo", Username: "user", Token: "token"}

	assert.Equal(t, "Gitea", registry.Name())
}

func TestGitea_Configured(t *testing.T) {
	tests := []struct {
		name       string
		registry   *Gitea
		configured bool
	}{
		{
			name:       "configured with repository",
			registry:   &Gitea{Repository: "org/repo"},
			configured: true,
		},
		{
			name:       "configured with registry only",
			registry:   &Gitea{Registry: "gitea.example.com"},
			configured: true,
		},
		{
			name:       "configured with both",
			registry:   &Gitea{Registry: "gitea.example.com", Repository: "org/repo"},
			configured: true,
		},
		{
			name:       "not configured",
			registry:   &Gitea{},
			configured: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.configured, tt.registry.Configured())
		})
	}
}

func TestGitea_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &Gitea{Registry: "gitea.example.com", Repository: "org/repo", Username: "user", Token: "secret-token-value"}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.Nil(t, err)
	assert.Equal(t, "user", client.Username)
	assert.Equal(t, "secret-token-value", client.Password)
	assert.Equal(t, "gitea.example.com", client.ServerAddress)
	logMock.Check(t, []string{
		"debug: Gitea login: registry=gitea.example.com, username=user, token=secr...alue[len=18]\n",
		"debug: Gitea login successful: Logged in\n",
	})
}

func TestGitea_LoginError(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &Gitea{Registry: "gitea.example.com", Username: "user", Token: "secret-token-value"}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.EqualError(t, err, "gitea registry login to gitea.example.com failed: invalid username/password")
	logMock.Check(t, []string{
		"debug: Gitea login: registry=gitea.example.com, username=user, token=secr...alue[len=18]\n",
		"error: Gitea login failed: invalid username/password\n",
	})
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{"empty token", "", "<empty>"},
		{"short token", "abc", "[len=3]"},
		{"8 char token", "12345678", "[len=8]"},
		{"long token", "secret-token-value", "secr...alue[len=18]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, maskToken(tt.token))
		})
	}
}

func TestGitea_GetAuthInfo(t *testing.T) {
	registry := &Gitea{Registry: "gitea.example.com", Repository: "org/repo", Username: "user", Token: "token"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6InVzZXIiLCJwYXNzd29yZCI6InRva2VuIiwic2VydmVyYWRkcmVzcyI6ImdpdGVhLmV4YW1wbGUuY29tIn0=", auth)
}

func TestGitea_GetAuthConfig(t *testing.T) {
	registry := &Gitea{Registry: "gitea.example.com", Repository: "org/repo", Username: "user", Token: "token"}
	auth := registry.GetAuthConfig()
	assert.Equal(t, "user", auth.Username)
	assert.Equal(t, "token", auth.Password)
	assert.Equal(t, "gitea.example.com", auth.ServerAddress)
}

func TestGitea_Create(t *testing.T) {
	registry := &Gitea{}
	err := registry.Create("repo")
	assert.Nil(t, err)
}

func TestGitea_RegistryUrl(t *testing.T) {
	tests := []struct {
		name        string
		registry    *Gitea
		expectedUrl string
	}{
		{
			name:        "with org/repo format",
			registry:    &Gitea{Registry: "gitea.example.com", Repository: "org/repo"},
			expectedUrl: "gitea.example.com/org",
		},
		{
			name:        "with user/repo format",
			registry:    &Gitea{Registry: "gitea.example.com", Repository: "user/myrepo"},
			expectedUrl: "gitea.example.com/user",
		},
		{
			name:        "with nested path org/subgroup/repo",
			registry:    &Gitea{Registry: "gitea.example.com", Repository: "org/subgroup/repo"},
			expectedUrl: "gitea.example.com/org/subgroup",
		},
		{
			name:        "with single name repository",
			registry:    &Gitea{Registry: "gitea.example.com", Repository: "singlename"},
			expectedUrl: "gitea.example.com/singlename",
		},
		{
			name:        "with registry only",
			registry:    &Gitea{Registry: "gitea.example.com"},
			expectedUrl: "gitea.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedUrl, tt.registry.RegistryUrl())
		})
	}
}
