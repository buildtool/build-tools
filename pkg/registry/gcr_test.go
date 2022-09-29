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

package registry

import (
	"fmt"
	"testing"

	"github.com/apex/log"
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg/docker"
)

func TestGcr_Name(t *testing.T) {
	registry := &GCR{
		Url:            "url",
		KeyFileContent: "a2V5ZmlsZSBjb250ZW50Cg==",
	}
	assert.Equal(t, "GCR", registry.Name())
}

func TestGcr_Configured_Missing_Url(t *testing.T) {
	registry := &GCR{
		KeyFileContent: "a2V5ZmlsZSBjb250ZW50Cg==",
	}

	assert.False(t, registry.Configured())
}

func TestGcr_Configured_Missing_Key(t *testing.T) {
	registry := &GCR{
		Url: "url",
	}

	assert.False(t, registry.Configured())
}

func TestGcr_Configured(t *testing.T) {
	registry := &GCR{
		Url:            "url",
		KeyFileContent: "a2V5ZmlsZSBjb250ZW50Cg==",
	}

	assert.True(t, registry.Configured())
}

func TestGcr_GetAuthConfig_Invalid_Base64Content(t *testing.T) {
	registry := &GCR{
		Url:            "url",
		KeyFileContent: "YWJjZA=====",
	}

	assert.Equal(t, types.AuthConfig{}, registry.GetAuthConfig())
}

func TestGcr_GetAuthInfo(t *testing.T) {
	registry := &GCR{
		Url:            "url",
		KeyFileContent: "a2V5ZmlsZSBjb250ZW50Cg==",
	}

	auth := registry.GetAuthInfo()
	assert.Equal(t, "eyJ1c2VybmFtZSI6Il9qc29uX2tleSIsInBhc3N3b3JkIjoia2V5ZmlsZSBjb250ZW50XG4ifQ==", auth)
}

func TestGcr_LoginFailed(t *testing.T) {
	client := &docker.MockDocker{LoginError: fmt.Errorf("invalid username/password")}
	registry := &GCR{
		Url:            "url",
		KeyFileContent: "a2V5ZmlsZSBjb250ZW50Cg==",
	}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.EqualError(t, err, "invalid username/password")
	logMock.Check(t, []string{})
}

func TestGcr_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &GCR{
		Url:            "url",
		KeyFileContent: "a2V5ZmlsZSBjb250ZW50Cg==",
	}
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	err := registry.Login(client)
	assert.Nil(t, err)
	assert.Equal(t, "_json_key", client.Username)
	assert.Equal(t, "keyfile content\n", client.Password)
	assert.Equal(t, "url", client.ServerAddress)
	logMock.Check(t, []string{"debug: Logged in\n"})
}
