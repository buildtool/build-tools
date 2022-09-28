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

package ci

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/buildtool/build-tools/pkg/vcs"
)

func TestGitlab_Name(t *testing.T) {
	ci := &Gitlab{}
	assert.Equal(t, "Gitlab", ci.Name())
}

func TestGitlab_BuildName(t *testing.T) {
	ci := &Gitlab{Common: &Common{}, CIBuildName: "Name"}

	assert.Equal(t, "name", ci.BuildName())
}

func TestGitlab_BuildName_Fallback(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	oldpwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldpwd) }()
	_ = os.Chdir(name)

	ci := &Gitlab{Common: &Common{}}

	assert.Equal(t, filepath.Base(name), ci.BuildName())
}

func TestGitlab_Branch(t *testing.T) {
	ci := &Gitlab{CIBranchName: "feature1"}

	assert.Equal(t, "feature1", ci.Branch())
}

func TestGitlab_Branch_Fallback(t *testing.T) {
	ci := &Gitlab{Common: &Common{VCS: vcs.NewMockVcs()}}

	assert.Equal(t, "fallback-branch", ci.Branch())
}

func TestGitlab_Commit(t *testing.T) {
	ci := &Gitlab{CICommit: "sha"}

	assert.Equal(t, "sha", ci.Commit())
}

func TestGitlab_Commit_Fallback(t *testing.T) {
	ci := &Gitlab{Common: &Common{VCS: vcs.NewMockVcs()}}

	assert.Equal(t, "fallback-sha", ci.Commit())
}
