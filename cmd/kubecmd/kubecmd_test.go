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

package main

import (
	"bytes"
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
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() {
		assert.NoError(t, os.RemoveAll(name))
	}()
	yaml := `
targets:
  dummy:
    context: missing
    namespace: none
`
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0o777)

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
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() {
		assert.NoError(t, os.RemoveAll(name))
	}()
	yaml := `
targets:
  dummy:
    context: local
    namespace: default
`
	_ = os.WriteFile(filepath.Join(name, ".buildtools.yaml"), []byte(yaml), 0o777)

	err := os.Chdir(name)
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"kubecmd", "dummy"}
	main()
	assert.Equal(t, "kubectl --context local --namespace default", out.(*bytes.Buffer).String())
}
