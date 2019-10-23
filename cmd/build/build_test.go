package main

import (
	"bytes"
	"github.com/sparetimecoders/build-tools/pkg"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestBuild(t *testing.T) {
	defer pkg.SetEnv("DOCKER_HOST", "abc-123")()
	exitFunc = func(code int) {
		assert.Equal(t, -1, code)
	}
	os.Args = []string{"build"}
	main()
}

func TestVersion(t *testing.T) {
	out = &bytes.Buffer{}
	exitFunc = func(code int) {
		assert.Equal(t, 0, code)
	}
	os.Args = []string{"build", "-version"}
	main()

	assert.Equal(t, "Version: dev, commit none, built at unknown\n", out.(*bytes.Buffer).String())
}
