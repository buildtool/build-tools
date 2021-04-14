package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/buildtool/build-tools/pkg"
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
	os.Args = []string{"build", "--version"}
	main()

	assert.Equal(t, "Version: dev, commit none, built at unknown\n", out.(*bytes.Buffer).String())
}

func TestArguments(t *testing.T) {
	out = &bytes.Buffer{}
	eout = &bytes.Buffer{}
	exitFunc = func(code int) {
		assert.Equal(t, -1, code)
	}
	os.Args = []string{"build", "--unknown"}
	main()

	assert.Equal(t, "build: error: unknown flag --unknown\n", eout.(*bytes.Buffer).String())
	assert.Contains(t, out.(*bytes.Buffer).String(), "Usage: build")
}
