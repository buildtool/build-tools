package service_setup

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var name string

func TestMain(m *testing.M) {
	tempDir := setup()
	code := m.Run()
	teardown(tempDir)
	os.Exit(code)
}

func setup() string {
	name, _ = ioutil.TempDir(os.TempDir(), "build-tools")
	_ = os.Chdir(name)

	return name
}

func teardown(tempDir string) {
	_ = os.RemoveAll(tempDir)
}

func TestSetup_NoArgs(t *testing.T) {
	out := bytes.Buffer{}

	exitCode := 0
	Setup(name, &out, func(code int) {
		exitCode = code
	})

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "\x1b[0mUsage: service-setup [options] <name>\n\nFor example \x1b[34m`service-setup --stack go gosvc`\x1b[39m would create a new repository and scaffold it as a Go-project\n\nOptions:\n\x1b[0m", out.String())
}

func TestSetup_NonExistingStack(t *testing.T) {
	out := bytes.Buffer{}

	exitCode := 0
	Setup(name, &out, func(code int) {
		exitCode = code
	}, "-s", "missing", "project")

	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[31mProvided stack does not exist yet. Available stacks are: \x1b[39m\x1b[97m\x1b[1m(go, none, scala)\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m", out.String())
}

func TestSetup_BrokenConfig(t *testing.T) {
	os.Clearenv()
	yaml := `ci: [] `
	file := filepath.Join(name, ".buildtools.yaml")
	_ = ioutil.WriteFile(file, []byte(yaml), 0777)
	defer func() { _ = os.Remove(file) }()

	out := bytes.Buffer{}

	exitCode := 0
	Setup(name, &out, func(code int) {
		exitCode = code
	}, "project")

	assert.Equal(t, -1, exitCode)
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s'\x1b[39m\x1b[0m\n\x1b[0m\x1b[31myaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig\x1b[39m\x1b[0m\n", file), out.String())
}

func TestSetup_NoVCS(t *testing.T) {
	out := bytes.Buffer{}

	exitCode := 0
	Setup(name, &out, func(code int) {
		exitCode = code
	}, "project")

	assert.Equal(t, -2, exitCode)
	assert.Equal(t, "\x1b[0m\x1b[31mno VCS configured\x1b[39m\x1b[0m\n", out.String())
}

//func TestSetup_BasicArgs(t *testing.T) {
//	yaml := `
//scaffold:
//  ci:
//    selected: gitlab
//  vcs:
//    selected: gitlab
//  registry: registry.gitlab.com/group
//`
//	filePath := filepath.Join(name, ".buildtools.yaml")
//	_ = ioutil.WriteFile(filePath, []byte(yaml), 0777)
//
//	out := bytes.Buffer{}
//
//	exitCode := 0
//	Setup(name, &out, func(code int) {
//		exitCode = code
//	}, "project")
//
//	assert.Equal(t, 0, exitCode)
//	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s'\x1b[39m\x1b[0m\n\x1b[0m\x1b[94mCreating new service \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m \x1b[94musing stack \x1b[39m\x1b[97m\x1b[1m'none'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating repository at \x1b[39m\x1b[97m\x1b[1m'none'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[32mCreated repository \x1b[39m\x1b[97m\x1b[1m''\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m\x1b[0m\x1b[94mCreating build pipeline for \x1b[39m\x1b[97m\x1b[1m'project'\x1b[0m\x1b[97m\x1b[39m\n\x1b[0m", filePath), out.String())
//}
