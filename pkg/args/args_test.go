// MIT License
//
// Copyright (c) 2018 buildtool
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package args

import (
	"encoding/base64"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/require"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg"
	"github.com/buildtool/build-tools/pkg/cli"
	"github.com/buildtool/build-tools/pkg/version"
)

func Test_Parse(t *testing.T) {
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)

	arguments := &struct {
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

	arguments := &struct {
		Globals
		Name string
	}{}
	err := ParseArgs("", []string{"--help"}, version.Info{
		Name:        "command",
		Description: "desc",
	}, arguments)
	require.Equal(t, err, ErrDone)
	logMock.Check(t, []string{
		"info: Usage: command [flags]\n",
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

	arguments := &struct {
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
	require.Equal(t, err, ErrDone)
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

	arguments := &struct {
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
	require.Equal(t, err, ErrDone)
	logMock.Check(t, []string{
		"debug: Parsing config from env: BUILDTOOLS_CONTENT\n",
		"info: Current config\nci: none\nvcs: Git\nregistry: {}" + yaml,
	})
}

func Test_Config_Error(t *testing.T) {
	yaml := `_`
	defer pkg.SetEnv("BUILDTOOLS_CONTENT", base64.StdEncoding.EncodeToString([]byte(yaml)))()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)

	arguments := &struct {
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
	require.Error(t, err)
	logMock.Check(t, []string{
		"debug: Parsing config from env: BUILDTOOLS_CONTENT\n",
		"info: name: error: yaml: unmarshal errors:\n",
		"info:                line 1: cannot unmarshal !!str `_` into config.Config\n",
	})
}

func Test_Verbose_Enabled(t *testing.T) {
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.InfoLevel)

	arguments := &struct {
		Globals
		Name string
	}{}
	_ = ParseArgs("", []string{"--verbose", "--name", "thename"}, version.Info{
		Name:        "name",
		Description: "desc",
		Version:     "version",
		Commit:      "commit",
		Date:        "date",
	}, arguments)
	logMock.Check(t, []string{})
	require.True(t, cli.Verbose(log.Log))
}

func Test_Verbose(t *testing.T) {
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.InfoLevel)

	arguments := &struct {
		Globals
		Name string
	}{}
	_ = ParseArgs("", []string{"--name", "thename"}, version.Info{
		Name:        "name",
		Description: "desc",
		Version:     "version",
		Commit:      "commit",
		Date:        "date",
	}, arguments)
	logMock.Check(t, []string{})
	require.False(t, cli.Verbose(log.Log))
}
