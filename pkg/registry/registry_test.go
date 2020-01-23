package registry

import (
	"bytes"
	"errors"
	"github.com/buildtool/build-tools/pkg/docker"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDockerRegistry_PushImage(t *testing.T) {
	registry := &Gitlab{}
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}
	client := &docker.MockDocker{PushError: errors.New("error")}

	err := registry.PushImage(client, "dummy", "unknown", out, eout)
	assert.EqualError(t, err, "error")
}
