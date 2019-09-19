package vcs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGitlab_Name(t *testing.T) {
	vcs := &Gitlab{}
	assert.Equal(t, "Gitlab", vcs.Name())
}
