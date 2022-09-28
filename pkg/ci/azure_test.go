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

func TestAzure_Name(t *testing.T) {
	ci := &Azure{}

	assert.Equal(t, "Azure", ci.Name())
}

func TestAzure_BuildName(t *testing.T) {
	ci := &Azure{Common: &Common{}, CIBuildName: "Name"}

	assert.Equal(t, "name", ci.BuildName())
}

func TestAzure_BuildName_Fallback(t *testing.T) {
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	oldpwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldpwd) }()
	_ = os.Chdir(name)

	ci := &Azure{Common: &Common{}}

	assert.Equal(t, filepath.Base(name), ci.BuildName())
}

func TestAzure_Branch(t *testing.T) {
	ci := &Azure{CIBranchName: "feature1"}

	assert.Equal(t, "feature1", ci.Branch())
}

func TestAzure_Branch_Fallback(t *testing.T) {
	ci := &Azure{Common: &Common{VCS: vcs.NewMockVcs()}}

	assert.Equal(t, "fallback-branch", ci.Branch())
}

func TestAzure_Commit(t *testing.T) {
	ci := &Azure{CICommit: "sha"}

	assert.Equal(t, "sha", ci.Commit())
}

func TestAzure_Commit_Fallback(t *testing.T) {
	ci := &Azure{Common: &Common{VCS: vcs.NewMockVcs()}}

	assert.Equal(t, "fallback-sha", ci.Commit())
}
