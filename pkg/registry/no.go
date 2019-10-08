package registry

import (
	"fmt"
	"gitlab.com/sparetimecoders/build-tools/pkg/docker"
	"io"
)

type NoDockerRegistry struct{}

func (n NoDockerRegistry) Configured() bool {
	return true
}

func (n NoDockerRegistry) Name() string {
	return "No docker registry"
}

func (n NoDockerRegistry) Login(client docker.Client, out io.Writer) error {
	_, _ = fmt.Fprintln(out, "Authentication not supported for registry")
	return nil
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

func (n NoDockerRegistry) PushImage(client docker.Client, auth, image string, out, eout io.Writer) error {
	return fmt.Errorf("push not supported by registry")
}

var _ Registry = &NoDockerRegistry{}
