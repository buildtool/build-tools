package vcs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGithub_Name(t *testing.T) {
	vcs := &Github{}
	assert.Equal(t, vcs.Name(), "Github")
}
