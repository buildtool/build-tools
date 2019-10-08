package registry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NoDockerRegistry_Name(t *testing.T) {
	assert.Equal(t, true, NoDockerRegistry{}.Configured())
}
