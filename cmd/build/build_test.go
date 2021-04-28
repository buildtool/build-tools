package main

import (
	"os"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg"
)

func TestBuild(t *testing.T) {
	logMock := mocks.New()
	handler = logMock
	defer pkg.SetEnv("DOCKER_HOST", "abc-123")()
	exitFunc = func(code int) {
		assert.Equal(t, -1, code)
	}
	os.Args = []string{"build"}
	main()
	logMock.Check(t, []string{"error: unable to parse docker host `abc-123`"})
}

func TestVersion(t *testing.T) {
	logMock := mocks.New()
	handler = logMock
	log.SetLevel(log.DebugLevel)
	exitFunc = func(code int) {
		assert.Equal(t, 0, code)
	}
	os.Args = []string{"build", "--version"}
	main()

	logMock.Check(t, []string{"info: Version: dev, commit none, built at unknown\n"})
}

func TestArguments(t *testing.T) {
	logMock := mocks.New()
	handler = logMock
	log.SetLevel(log.DebugLevel)
	exitFunc = func(code int) {
		assert.Equal(t, -1, code)
	}
	os.Args = []string{"build", "--unknown"}
	main()

	logMock.Check(t, []string{
		"info: build: error: unknown flag --unknown\n",
	})
}
