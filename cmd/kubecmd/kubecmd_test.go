package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestVersion(t *testing.T) {
	out = &bytes.Buffer{}
	version = "1.0.0"
	commit = "67d2fcf276fcd9cf743ad4be9a9ef5828adc082f"
	exitFunc = func(code int) {
		assert.Equal(t, 0, code)
	}
	os.Args = []string{"build", "-version"}
	main()

	assert.Equal(t, "Version: 1.0.0, commit 67d2fcf276fcd9cf743ad4be9a9ef5828adc082f, built at unknown\n", out.(*bytes.Buffer).String())
}

func TestKubecmd_MissingArguments(t *testing.T) {
	os.Args = []string{"kubecmd"}
	main()
}

func TestKubecmd_NoOptions(t *testing.T) {
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	yaml := `
environments:
  dummy:
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

func TestKubecmd_Output(t *testing.T) {
	out = &bytes.Buffer{}
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	yaml := `
environments:
  dummy:
    context: local
    namespace: default
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer os.Chdir(oldPwd)

	os.Args = []string{"deploy", "dummy"}
	main()
	assert.Equal(t, "kubectl --context local --namespace default", out.(*bytes.Buffer).String())
}
