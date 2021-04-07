package registry

import (
	"bytes"
	"fmt"
	"github.com/buildtool/build-tools/pkg/docker"
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	"testing"
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
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.EqualError(t, err, "invalid username/password")
	assert.Equal(t, "", out.String())
}

func TestGcr_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &GCR{
		Url:            "url",
		KeyFileContent: "a2V5ZmlsZSBjb250ZW50Cg==",
	}
	out := &bytes.Buffer{}
	err := registry.Login(client, out)
	assert.Nil(t, err)
	assert.Equal(t, "_json_key", client.Username)
	assert.Equal(t, "keyfile content\n", client.Password)
	assert.Equal(t, "url", client.ServerAddress)
	assert.Equal(t, "Logged in\n", out.String())
}
