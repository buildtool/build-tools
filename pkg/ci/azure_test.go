package ci

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAzure_Name(t *testing.T) {
	ci := &Azure{}

	assert.Equal(t, "Azure", ci.Name())
}
