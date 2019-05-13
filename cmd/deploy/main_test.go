package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestDeploy_BrokenConfig(t *testing.T) {
	os.Clearenv()
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	yaml := `ci: []
`
	_ = ioutil.WriteFile(filepath.Join(name, "buildtools.yaml"), []byte(yaml), 0777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer os.Chdir(oldPwd)

	os.Args = []string{"deploy", "dummy"}
	main()
}

func TestDeploy_MissingEnvironment(t *testing.T) {
	os.Clearenv()
	os.Args = []string{"deploy", "dummy"}
	main()
}

func TestDeploy_NoCI(t *testing.T) {
	os.Clearenv()
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	yaml := `
environments:
  - name: dummy
    context: missing
    namespace: none
`
	_ = ioutil.WriteFile(filepath.Join(name, "buildtools.yaml"), []byte(yaml), 0777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer os.Chdir(oldPwd)

	os.Args = []string{"deploy", "dummy"}
	main()
}

func TestDeploy_Fluff(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("CI_COMMIT_SHA", "abc123")
	_ = os.Setenv("CI_PROJECT_NAME", "dummy")
	_ = os.Setenv("CI_COMMIT_REF_NAME", "master")
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	yaml := `
environments:
  - name: dummy
    context: missing
    namespace: none
`
	_ = ioutil.WriteFile(filepath.Join(name, "buildtools.yaml"), []byte(yaml), 0777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer os.Chdir(oldPwd)

	os.Args = []string{"deploy", "dummy"}
	main()
}
