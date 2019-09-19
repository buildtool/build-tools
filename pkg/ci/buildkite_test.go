package ci

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestName_Buildkite(t *testing.T) {
	ci := &Buildkite{}
	assert.Equal(t, "Buildkite", ci.Name())
}
