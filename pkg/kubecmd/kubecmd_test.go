package kubecmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

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
	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	cmd := Kubecmd(".", info)
	assert.Nil(t, cmd)
	//assert.Contains(t, "Usage: kubecmd <target>")
}

func TestKubecmd_BrokenConfig(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `ci: []
`
	filePath := filepath.Join(name, ".buildtools.yaml")
	_ = os.WriteFile(filePath, []byte(yaml), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	cmd := Kubecmd(name, info, "dummy")
	assert.Nil(t, cmd)
	logMock.Check(t, []string{
		fmt.Sprintf("debug: Parsing config from file: <green>'%s'</green>\n", filePath),
		"error: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!seq into config.CIConfig"})
}

func TestKubecmd_MissingTarget(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := ``
	filePath := filepath.Join(name, ".buildtools.yaml")
	_ = os.WriteFile(filePath, []byte(yaml), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	cmd := Kubecmd(name, info, "dummy")
	assert.Nil(t, cmd)
	logMock.Check(t, []string{fmt.Sprintf(
		"debug: Parsing config from file: <green>'%s'</green>\n", filePath),
		"error: no target matching dummy found"})
}

func TestKubecmd_NoOptions(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  dummy:
    context: missing
    namespace: none
`
	filePath := filepath.Join(name, ".buildtools.yaml")
	_ = os.WriteFile(filePath, []byte(yaml), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	cmd := Kubecmd(name, info, "dummy")
	assert.Equal(t, "kubectl --context missing --namespace none", *cmd)
	logMock.Check(t, []string{fmt.Sprintf("debug: Parsing config from file: <green>'%s'</green>\n", filePath)})
}

func TestKubecmd_ContextAndNamespaceSpecified(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	yaml := `
targets:
  dummy:
    context: missing
    namespace: none
`
	filePath := filepath.Join(name, ".buildtools.yaml")
	_ = os.WriteFile(filePath, []byte(yaml), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	cmd := Kubecmd(name, info, "-c", "other", "-n", "dev", "dummy")
	assert.Equal(t, "kubectl --context other --namespace dev", *cmd)
	logMock.Check(t, []string{fmt.Sprintf("debug: Parsing config from file: <green>'%s'</green>\n", filePath)})
}
