package vcs

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNo_Identify(t *testing.T) {
	vcs := &no{}

	out := &bytes.Buffer{}

	assert.True(t, vcs.Identify("test", out))
	assert.Equal(t, "", out.String())
}

func TestNo_Name(t *testing.T) {
	vcs := &no{}

	assert.Equal(t, "none", vcs.Name())
}
