package args

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/buildtool/build-tools/pkg"
	"github.com/buildtool/build-tools/pkg/version"
)

func Test_Parse(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}

	var arguments = &struct {
		Globals
		Name string
	}{}
	err := ParseArgs("", out, eout, []string{"--name", "thename"}, version.Info{}, arguments)
	require.NoError(t, err)
	require.Equal(t, "thename", arguments.Name)
}

func Test_Help(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}

	var arguments = &struct {
		Globals
		Name string
	}{}
	err := ParseArgs("", out, eout, []string{"--help"}, version.Info{
		Name:        "command",
		Description: "desc",
	}, arguments)
	require.Equal(t, err, Done)
	require.Equal(t, "Usage: command\n\ndesc\n\nFlags:\n  -h, --help           Show context-sensitive help.\n      --version        Print args information and quit\n  -v, --verbose        Enable verbose mode\n      --config\n      --name=STRING\n", out.String())
}

func Test_Version(t *testing.T) {
	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}

	var arguments = &struct {
		Globals
		Name string
	}{}
	err := ParseArgs("", out, eout, []string{"--version", "--name", "thename"}, version.Info{
		Name:        "name",
		Description: "desc",
		Version:     "version",
		Commit:      "commit",
		Date:        "date",
	}, arguments)
	require.Equal(t, err, Done)
	require.Equal(t, "Version: version, commit commit, built at date\n", out.String())
}

func Test_Config(t *testing.T) {
	yaml := `
targets:
    local:
        context: docker-desktop
`
	defer pkg.SetEnv("BUILDTOOLS_CONTENT", base64.StdEncoding.EncodeToString([]byte(yaml)))()

	out := &bytes.Buffer{}
	eout := &bytes.Buffer{}

	var arguments = &struct {
		Globals
		Name string
	}{}
	err := ParseArgs("", out, eout, []string{"--config", "--name", "thename"}, version.Info{
		Name:        "name",
		Description: "desc",
		Version:     "version",
		Commit:      "commit",
		Date:        "date",
	}, arguments)
	require.Equal(t, err, Done)
	require.Equal(t, "Parsing config from env: BUILDTOOLS_CONTENT\nCurrent config\nci: none\nvcs: Git\nregistry: {}"+yaml, out.String())
}
