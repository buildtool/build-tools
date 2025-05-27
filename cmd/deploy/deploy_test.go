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

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg"
)

func TestVersion(t *testing.T) {
	logMock := mocks.New()
	handler = logMock
	log.SetLevel(log.DebugLevel)
	version = "1.0.0"
	exitFunc = func(code int) {
		assert.Equal(t, 0, code)
	}
	os.Args = []string{"deploy", "--version"}
	main()

	logMock.Check(t, []string{"info: Version: 1.0.0, commit none, built at unknown\n"})
}

func TestDeploy_BrokenConfig(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -1, code)
	}

	oldPwd, _ := os.Getwd()
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `ci: []
`
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0o777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy", "dummy"}
	main()
}

func TestDeploy_MissingTargetAndContext(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -5, code)
	}

	oldPwd, _ := os.Getwd()
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy", "dummy"}
	main()
}

func TestDeploy_NoCI(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -3, code)
	}
	defer pkg.UnsetGithubEnvironment()()

	oldPwd, _ := os.Getwd()
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  dummy:
    context: missing
    namespace: none
`
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0o777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy", "dummy"}
	main()
}

func TestDeploy_NoEnv(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -1, code)
	}

	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "dummy")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	oldPwd, _ := os.Getwd()
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  - name: dummy
    context: missing
    namespace: none
`
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0o777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy"}
	main()
}

func TestDeploy_NoOptions(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -4, code)
	}

	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "dummy")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	oldPwd, _ := os.Getwd()
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  dummy:
    context: missing
    namespace: none
`
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0o777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy", "dummy"}
	main()
}

func TestDeploy_ContextAndNamespaceSpecified(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -4, code)
	}

	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "dummy")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	oldPwd, _ := os.Getwd()
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  dummy:
    context: missing
    namespace: none
`
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0o777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy", "--context", "other", "--namespace", "dev", "dummy"}
	main()
}

func TestDeploy_Timeout(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -4, code)
	}

	defer pkg.SetEnv("CI_COMMIT_SHA", "abc123")()
	defer pkg.SetEnv("CI_PROJECT_NAME", "dummy")()
	defer pkg.SetEnv("CI_COMMIT_REF_NAME", "master")()
	oldPwd, _ := os.Getwd()
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  dummy:
    context: missing
    namespace: none
`
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0o777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy", "--timeout", "20s", "dummy"}
	main()
}

func TestDeploy_Tag(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -4, code)
	}
	oldPwd, _ := os.Getwd()
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  dummy:
    context: missing
    namespace: none
`
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0o777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy", "--tag", "123", "dummy"}
	main()
}
