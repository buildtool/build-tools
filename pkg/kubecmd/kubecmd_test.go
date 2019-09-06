package kubecmd

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestKubecmd_MissingArgumentsPrintUsageToOutWriter(t *testing.T) {
	out := bytes.Buffer{}
	cmd := Kubecmd(".", &out)
	assert.Nil(t, cmd)
	assert.Equal(t, "Usage: deploy [options] <environment>\n\nFor example `deploy --context test-cluster --namespace test prod` would deploy to namsepace `test` in the `test-cluster` but assuming to use the `prod` configuration files (if present)\n\nOptions:\n", out.String())
}

func TestKubecmd_BrokenConfig(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `ci: []
`
	filePath := filepath.Join(name, ".buildtools.yaml")
	_ = ioutil.WriteFile(filePath, []byte(yaml), 0777)

	out := bytes.Buffer{}
	cmd := Kubecmd(name, &out, "dummy")
	assert.Nil(t, cmd)
	assert.Equal(t, fmt.Sprintf("Parsing config from file: '%s'\nyaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig\n", filePath), out.String())
}

func TestKubecmd_MissingEnvironment(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := ``
	filePath := filepath.Join(name, ".buildtools.yaml")
	_ = ioutil.WriteFile(filePath, []byte(yaml), 0777)

	out := bytes.Buffer{}
	cmd := Kubecmd(name, &out, "dummy")
	assert.Nil(t, cmd)
	assert.Equal(t, fmt.Sprintf("Parsing config from file: '%s'\nno environment matching dummy found\n", filePath), out.String())
}

func TestKubecmd_NoOptions(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
environments:
  - name: dummy
    context: missing
    namespace: none
`
	filePath := filepath.Join(name, ".buildtools.yaml")
	_ = ioutil.WriteFile(filePath, []byte(yaml), 0777)

	out := bytes.Buffer{}
	cmd := Kubecmd(name, &out, "dummy")
	assert.Equal(t, "kubectl --context missing --namespace none", *cmd)
	assert.Equal(t, fmt.Sprintf("Parsing config from file: '%s'\n", filePath), out.String())
}

func TestKubecmd_ContextAndNamespaceSpecified(t *testing.T) {
	name, _ := ioutil.TempDir(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
environments:
  - name: dummy
    context: missing
    namespace: none
`
	filePath := filepath.Join(name, ".buildtools.yaml")
	_ = ioutil.WriteFile(filePath, []byte(yaml), 0777)

	out := bytes.Buffer{}
	cmd := Kubecmd(name, &out, "-c", "other", "-n", "dev", "dummy")
	assert.Equal(t, "kubectl --context other --namespace dev", *cmd)
	assert.Equal(t, fmt.Sprintf("Parsing config from file: '%s'\n", filePath), out.String())
}
