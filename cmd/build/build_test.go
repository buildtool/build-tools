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
		"info: Usage: build\n",
		"info: \n",
		"info: performs a docker build and tags the resulting image\n",
		"info: \n",
		"info: Flags:\n",
		"info:   -h, --help                       Show context-sensitive help.\n",
		"info:       --version                    Print args information and exit\n",
		"info:   -v, --verbose                    Enable verbose mode\n",
		"info:       --config                     Print parsed config and exit\n",
		"info:   -f, --file=\"Dockerfile\"          name of the Dockerfile to use.\n",
		"info:       --build-arg=BUILD-ARG,...    additional docker build-args to use, see\n",
		"info:                                    https://docs.docker.com/engine/reference/commandline/build/\n",
		"info:                                    for more information.\n",
		"info:       --no-login                   disable login to docker registry\n",
		"info:       --no-pull                    disable pulling latest from docker registry\n",
		"info: \n",
		"info: build: error: unknown flag --unknown\n",
	})
}
