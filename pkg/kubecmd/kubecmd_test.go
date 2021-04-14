package kubecmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/buildtool/build-tools/pkg/version"
)

var info = version.Info{
	Name:        "kubecmd",
	Description: "TODO",
	Version:     "1.0",
	Commit:      "c3ac9dbe16fee57d1f5ad04f6dbe3d5c13efb299",
	Date:        "2021-04-09T06:43:59Z",
}

func TestKubecmd_MissingArgumentsPrintUsageToOutWriter(t *testing.T) {
	eout := bytes.Buffer{}
	out := bytes.Buffer{}
	cmd := Kubecmd(".", &out, &eout, info)
	assert.Nil(t, cmd)
	assert.Contains(t, out.String(), "Usage: kubecmd <target>")
}

func TestKubecmd_BrokenConfig(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `ci: []
`
	filePath := filepath.Join(name, ".buildtools.yaml")
	_ = ioutil.WriteFile(filePath, []byte(yaml), 0777)

	eout := bytes.Buffer{}
	out := bytes.Buffer{}
	cmd := Kubecmd(name, &out, &eout, info, "dummy")
	assert.Nil(t, cmd)
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s'\x1b[39m\x1b[0m\nyaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig\n", filePath), out.String())
}

func TestKubecmd_MissingTarget(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := ``
	filePath := filepath.Join(name, ".buildtools.yaml")
	_ = ioutil.WriteFile(filePath, []byte(yaml), 0777)

	eout := bytes.Buffer{}
	out := bytes.Buffer{}
	cmd := Kubecmd(name, &out, &eout, info, "dummy")
	assert.Nil(t, cmd)
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s'\x1b[39m\x1b[0m\nno target matching dummy found\n", filePath), out.String())
}

func TestKubecmd_NoOptions(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  dummy:
    context: missing
    namespace: none
`
	filePath := filepath.Join(name, ".buildtools.yaml")
	_ = ioutil.WriteFile(filePath, []byte(yaml), 0777)

	eout := bytes.Buffer{}
	out := bytes.Buffer{}
	cmd := Kubecmd(name, &out, &eout, info, "dummy")
	assert.Equal(t, "kubectl --context missing --namespace none", *cmd)
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s'\x1b[39m\x1b[0m\n", filePath), out.String())
}

func TestKubecmd_ContextAndNamespaceSpecified(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  dummy:
    context: missing
    namespace: none
`
	filePath := filepath.Join(name, ".buildtools.yaml")
	_ = ioutil.WriteFile(filePath, []byte(yaml), 0777)

	eout := bytes.Buffer{}
	out := bytes.Buffer{}
	cmd := Kubecmd(name, &out, &eout, info, "-c", "other", "-n", "dev", "dummy")
	assert.Equal(t, "kubectl --context other --namespace dev", *cmd)
	assert.Equal(t, fmt.Sprintf("\x1b[0mParsing config from file: \x1b[32m'%s'\x1b[39m\x1b[0m\n", filePath), out.String())
}
