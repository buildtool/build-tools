package registry

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"os"
	"testing"
)

func TestEcr_Identify(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("ECR_URL", "url")
	_ = os.Setenv("ECR_REGION", "region")

	registry := Identify()
	assert.NotNil(t, registry)
	assert.Equal(t, "url", registry.RegistryUrl())
}

func TestEcr_LoginSuccess(t *testing.T) {
	client := &docker.MockDocker{}
	registry := &ecr{url: "ecr-url", region: "eu-west-1"}
	err := registry.Login(client)
	assert.Nil(t, err)
	// TODO: Fix when correct implementation is in place
	assert.Equal(t, "", client.Username)
	assert.Equal(t, "", client.Password)
	assert.Equal(t, "", client.ServerAddress)
}

func TestEcr_GetAuthInfo(t *testing.T) {
	registry := &ecr{url: "ecr-url", region: "eu-west-1"}
	auth := registry.GetAuthInfo()
	assert.Equal(t, "", auth)
}

func TestEcr_Create(t *testing.T) {
	registry := &ecr{}
	err := registry.Create()
	assert.Nil(t, err)
}
