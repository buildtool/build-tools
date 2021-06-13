package registry

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/buildtool/build-tools/pkg/docker"
)

func TestDockerRegistry_PushImage(t *testing.T) {
	registry := &Gitlab{}
	client := &docker.MockDocker{PushError: errors.New("error")}

	err := registry.PushImage(client, "dummy", "unknown")
	assert.EqualError(t, err, "error")
}
