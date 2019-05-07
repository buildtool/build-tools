package registry

import (
	"github.com/stretchr/testify/assert"
	docker2 "gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"os"
	"testing"
)

func TestIdentify_Ecr(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("ECR_URL", "url")
	_ = os.Setenv("ECR_REGION", "region")

	docker := &docker2.MockDocker{}
	result := Identify()
	assert.NotNil(t, result)
	assert.Equal(t, "url", result.RegistryUrl())
	result.Login(docker)
	// TODO: Fix when correct implementation is in place
	assert.Equal(t, "", docker.Username)
	assert.Equal(t, "", docker.Password)
	assert.Equal(t, "", docker.ServerAddress)
}
