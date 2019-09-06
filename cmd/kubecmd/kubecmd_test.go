package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestKubecmd_MissingArguments(t *testing.T) {
	os.Args = []string{"deploy"}
	main()
}

func TestKubecmd_BrokenConfig(t *testing.T) {
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	yaml := `ci: []
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer os.Chdir(oldPwd)

	os.Args = []string{"kubecmd", "dummy"}
	main()
}

func TestKubecmd_MissingEnvironment(t *testing.T) {
	os.Args = []string{"deploy", "dummy"}
	main()
}

func TestKubecmd_NoOptions(t *testing.T) {
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	yaml := `
environments:
  - name: dummy
    context: missing
    namespace: none
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer os.Chdir(oldPwd)

	os.Args = []string{"deploy", "dummy"}
	main()
}

func TestKubecmd_ContextAndNamespaceSpecified(t *testing.T) {
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	yaml := `
environments:
  - name: dummy
    context: missing
    namespace: none
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer os.Chdir(oldPwd)

	os.Args = []string{"deploy", "-c", "other", "-n", "dev", "dummy"}
	main()
}
