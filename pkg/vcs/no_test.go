package vcs

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNo_Identify(t *testing.T) {
	vcs := &no{}

	out := &bytes.Buffer{}

	assert.True(t, vcs.Identify("test"))
	assert.Equal(t, "", out.String())
}

func TestNo_Name(t *testing.T) {
	vcs := &no{}

	assert.Equal(t, "none", vcs.Name())
}
