package registry

import (
	"fmt"

	"github.com/apex/log"
	"github.com/docker/docker/api/types"

	"github.com/buildtool/build-tools/pkg/docker"
)

type NoDockerRegistry struct{}

func (n NoDockerRegistry) Configured() bool {
	return true
}

func (n NoDockerRegistry) Name() string {
	return "No docker registry"
}

func (n NoDockerRegistry) Login(client docker.Client) error {
	log.Debugf("Authentication <yellow>not supported</yellow> for registry <green>%s</green>\n", n.Name())
	return nil
}

func (n NoDockerRegistry) GetAuthConfig() types.AuthConfig {
	return types.AuthConfig{}
}

func (n NoDockerRegistry) GetAuthInfo() string {
	return ""
}

func (n NoDockerRegistry) RegistryUrl() string {
	return "noregistry"
}

func (n NoDockerRegistry) Create(repository string) error {
	return nil
}

func (n NoDockerRegistry) PushImage(client docker.Client, auth, image string) error {
	return fmt.Errorf("push not supported by registry")
}

var _ Registry = &NoDockerRegistry{}
