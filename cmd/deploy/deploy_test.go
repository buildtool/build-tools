package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestDeploy_BrokenConfig(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -1, code)
	}

	os.Clearenv()
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

func TestDeploy_MissingEnvironment(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -0, code)
	}

	os.Clearenv()
	os.Args = []string{"deploy", "dummy"}
	main()
}

func TestDeploy_NoCI(t *testing.T) {
	exitFunc = func(code int) {
		assert.Equal(t, -2, code)
	}

	os.Clearenv()
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
environments:
  - name: dummy
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
		assert.Equal(t, -0, code)
	}

	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "dummy")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "master")
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
environments:
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
		assert.Equal(t, -2, code)
	}

	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "dummy")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "master")
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
environments:
  - name: dummy
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
		assert.Equal(t, -2, code)
	}

	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "dummy")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "master")
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
environments:
  - name: dummy
    context: missing
    namespace: none
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"deploy", "-c", "other", "-n", "dev", "dummy"}
	main()
}
