package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NoDockerRegistry_Name(t *testing.T) {
	assert.Equal(t, true, NoDockerRegistry{}.Configured())
}
