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

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/assert"
	mocks "gitlab.com/unboundsoftware/apex-mocks"

	"github.com/buildtool/build-tools/pkg/vcs"
)

func TestIdentify(t *testing.T) {
	os.Clearenv()

	dir, _ := os.MkdirTemp("", "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	result := vcs.Identify(dir)
	assert.NotNil(t, result)
	assert.Equal(t, "none", result.Name())
	assert.Equal(t, "", result.Commit())
	assert.Equal(t, "", result.Branch())
	logMock.Check(t, []string{})
}

func TestGit_Identify(t *testing.T) {
	dir, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	hash, _ := InitRepoWithCommit(dir)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	result := vcs.Identify(dir)
	assert.NotNil(t, result)
	assert.Equal(t, "Git", result.Name())
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "master", result.Branch())
	logMock.Check(t, []string{})
}

func TestGit_Identify_Subdirectory(t *testing.T) {
	dir, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	hash, _ := InitRepoWithCommit(dir)

	subdir := filepath.Join(dir, "subdir")
	_ = os.Mkdir(subdir, 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	result := vcs.Identify(subdir)
	assert.NotNil(t, result)
	assert.Equal(t, "Git", result.Name())
	assert.Equal(t, hash.String(), result.Commit())
	assert.Equal(t, "master", result.Branch())
	logMock.Check(t, []string{})
}

func TestGit_MissingRepo(t *testing.T) {
	dir, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	_ = os.Mkdir(filepath.Join(dir, ".git"), 0777)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	result := vcs.Identify(dir)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.Commit())
	assert.Equal(t, "", result.Branch())
	logMock.Check(t, []string{})
}

func TestGit_NoCommit(t *testing.T) {
	dir, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(dir) }()

	InitRepo(dir)

	logMock := mocks.New()
	log.SetHandler(logMock)
	log.SetLevel(log.DebugLevel)
	result := vcs.Identify(dir)
	assert.NotNil(t, result)
	assert.Equal(t, "", result.Commit())
	assert.Equal(t, "", result.Branch())
	logMock.Check(t, []string{"debug: Unable to fetch head: reference not found\n"})
}
