package vcs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAzureVCS_Name(t *testing.T) {
	vcs := &Azure{}

	assert.Equal(t, "Azure", vcs.Name())
}
