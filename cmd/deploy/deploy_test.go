package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/buildtool/build-tools/pkg"
)

func TestVersion(t *testing.T) {
	out = &bytes.Buffer{}
	version = "1.0.0"
	exitFunc = func(code int) {
		assert.Equal(t, 0, code)
	}
	os.Args = []string{"deploy", "--version"}
	main()

	assert.Equal(t, "Version: 1.0.0, commit none, built at unknown\n", out.(*bytes.Buffer).String())
}

func TestDeploy_BrokenConfig(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -1, code)
	}

	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `ci: []
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

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
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
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

	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  dummy:
    context: missing
    namespace: none
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

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
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  - name: dummy
    context: missing
    namespace: none
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

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
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  dummy:
    context: missing
    namespace: none
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

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
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  dummy:
    context: missing
    namespace: none
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

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
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  dummy:
    context: missing
    namespace: none
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

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
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  dummy:
    context: missing
    namespace: none
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy", "--tag", "123", "dummy"}
	main()
}
