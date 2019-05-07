package registry

import (
	"github.com/stretchr/testify/assert"
	docker2 "gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"os"
	"testing"
)

func TestIdentify_Gitlab(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_REGISTRY_IMAGE", "image")
	_ = os.Setenv("CI_REGISTRY", "registry.gitlab.com/group/image")
	_ = os.Setenv("CI_BUILD_TOKEN", "token")

	docker := &docker2.MockDocker{}
	result := Identify()
	assert.NotNil(t, result)
	assert.Equal(t, "registry.gitlab.com/group", result.RegistryUrl())
	result.Login(docker)
	assert.Equal(t, "gitlab-ci-token", docker.Username)
	assert.Equal(t, "token", docker.Password)
	assert.Equal(t, "registry.gitlab.com/group", docker.ServerAddress)
}
