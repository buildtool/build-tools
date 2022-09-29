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
