package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestBuild(t *testing.T) {
	_ = os.Setenv("DOCKER_HOST", "abc-123")
	exitFunc = func(code int) {
		assert.Equal(t, -1, code)
	}
	os.Args = []string{"build"}
	main()
}
