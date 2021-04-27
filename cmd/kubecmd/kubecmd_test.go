package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"
)

func TestVersion(t *testing.T) {
	logMock := mocks.New()
	handler = logMock
	log.SetLevel(log.DebugLevel)
	version = "1.0.0"
	commit = "67d2fcf276fcd9cf743ad4be9a9ef5828adc082f"
	os.Args = []string{"kubecmd", "--version"}
	main()

	logMock.Check(t, []string{"info: Version: 1.0.0, commit 67d2fcf276fcd9cf743ad4be9a9ef5828adc082f, built at unknown\n"})
}

func TestKubecmd_MissingArguments(t *testing.T) {
	os.Args = []string{"kubecmd"}
	main()
}

func TestKubecmd_NoOptions(t *testing.T) {
	out = &bytes.Buffer{}
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
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

	os.Args = []string{"kubecmd", "dummy"}
	main()
	assert.Equal(t, "kubectl --context missing --namespace none", out.(*bytes.Buffer).String())

}

func TestKubecmd_Output(t *testing.T) {
	out = &bytes.Buffer{}
	oldPwd, _ := os.Getwd()
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer os.RemoveAll(name)
	yaml := `
targets:
  dummy:
    context: local
    namespace: default
`
	_ = ioutil.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"kubecmd", "dummy"}
	main()
	assert.Equal(t, "kubectl --context local --namespace default", out.(*bytes.Buffer).String())
}
