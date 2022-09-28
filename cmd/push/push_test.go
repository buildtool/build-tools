// MIT License
//
// Copyright (c) 2021 buildtool
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
	"fmt"
	"os"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"
)

func TestPush(t *testing.T) {
	logMock := mocks.New()
	handler = logMock
	os.Clearenv()
	exitFunc = func(code int) {
		assert.Equal(t, -5, code)
	}

	oldPwd, _ := os.Getwd()
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()

	err := os.Chdir(name)
	assert.NoError(t, err)
	workDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldPwd) }()

	os.Args = []string{"push"}
	main()
	logMock.Check(t, []string{fmt.Sprintf("error: <red>open %s/Dockerfile: no such file or directory</red>", workDir)})
}

func TestVersion(t *testing.T) {
	logMock := mocks.New()
	handler = logMock
	log.SetLevel(log.DebugLevel)
	version = "1.0.0"
	commit = "67d2fcf276fcd9cf743ad4be9a9ef5828adc082f"
	date = "2006-01-02T15:04:05Z07:00"
	exitFunc = func(code int) {
		assert.Equal(t, 0, code)
	}
	os.Args = []string{"push", "--version"}
	main()

	logMock.Check(t, []string{"info: Version: 1.0.0, commit 67d2fcf276fcd9cf743ad4be9a9ef5828adc082f, built at 2006-01-02T15:04:05Z07:00\n"})
}
