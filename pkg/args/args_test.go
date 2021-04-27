package args

import (
	"encoding/base64"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/require"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg"
	"github.com/buildtool/build-tools/pkg/version"
)

func Test_Parse(t *testing.T) {
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)

	var arguments = &struct {
		Globals
		Name string
	}{}
	err := ParseArgs("", []string{"--name", "thename"}, version.Info{}, arguments)
	require.NoError(t, err)
	require.Equal(t, "thename", arguments.Name)
}

func Test_Help(t *testing.T) {
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)

	var arguments = &struct {
		Globals
		Name string
	}{}
	err := ParseArgs("", []string{"--help"}, version.Info{
		Name:        "command",
		Description: "desc",
	}, arguments)
	require.Equal(t, err, Done)
	logMock.Check(t, []string{
		"info: Usage: command\n",
		"info: \n",
		"info: desc\n",
		"info: \n",
		"info: Flags:\n",
		"info:   -h, --help           Show context-sensitive help.\n",
		"info:       --version        Print args information and exit\n",
		"info:   -v, --verbose        Enable verbose mode\n",
		"info:       --config         Print parsed config and exit\n",
		"info:       --name=STRING\n",
	})
}

func Test_Version(t *testing.T) {
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)

	var arguments = &struct {
		Globals
		Name string
	}{}
	err := ParseArgs("", []string{"--version", "--name", "thename"}, version.Info{
		Name:        "name",
		Description: "desc",
		Version:     "version",
		Commit:      "commit",
		Date:        "date",
	}, arguments)
	require.Equal(t, err, Done)
	logMock.Check(t, []string{"info: Version: version, commit commit, built at date\n"})
}

func Test_Config(t *testing.T) {
	yaml := `
targets:
    local:
        context: docker-desktop
`
	defer pkg.SetEnv("BUILDTOOLS_CONTENT", base64.StdEncoding.EncodeToString([]byte(yaml)))()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)

	var arguments = &struct {
		Globals
		Name string
	}{}
	err := ParseArgs("", []string{"--config", "--name", "thename"}, version.Info{
		Name:        "name",
		Description: "desc",
		Version:     "version",
		Commit:      "commit",
		Date:        "date",
	}, arguments)
	require.Equal(t, err, Done)
	logMock.Check(t, []string{"debug: Parsing config from env: BUILDTOOLS_CONTENT\n",
		"info: Current config\nci: none\nvcs: Git\nregistry: {}" + yaml})
}
