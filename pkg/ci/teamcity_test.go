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

package ci

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/buildtool/build-tools/pkg/vcs"
)

func TestTeamCityCI_Name(t *testing.T) {
	ci := &TeamCity{}

	assert.Equal(t, "TeamCity", ci.Name())
}

func TestTeamCityCI_BranchReplaceSlash(t *testing.T) {
	ci := &TeamCity{CIBranchName: "refs/heads/feature1"}

	assert.Equal(t, "refs_heads_feature1", ci.BranchReplaceSlash())
}

func TestTeamCityCI_BranchReplaceSlash_VCS_Fallback(t *testing.T) {
	ci := &TeamCity{Common: &Common{VCS: vcs.NewMockVcsWithBranch("refs/heads/feature1")}}

	assert.Equal(t, "refs_heads_feature1", ci.BranchReplaceSlash())
}

func TestTeamCityCI_BuildName(t *testing.T) {
	ci := &TeamCity{Common: &Common{}, CIBuildName: "project"}

	assert.Equal(t, "project", ci.BuildName())
}

func TestTeamCityCI_BuildName_VCS_Fallback(t *testing.T) {
	oldpwd, _ := os.Getwd()
	name, _ := os.MkdirTemp(os.TempDir(), "build-tools")
	defer func() { _ = os.RemoveAll(name) }()
	_ = os.Chdir(name)
	defer func() { _ = os.Chdir(oldpwd) }()

	ci := &TeamCity{Common: &Common{VCS: vcs.NewMockVcs()}}

	assert.Equal(t, filepath.Base(name), ci.BuildName())
}

func TestTeamCityCI_Commit(t *testing.T) {
	ci := &TeamCity{CICommit: "abc123"}

	assert.Equal(t, "abc123", ci.Commit())
}

func TestTeamCityCI_Commit_VCS_Fallback(t *testing.T) {
	ci := &TeamCity{Common: &Common{VCS: vcs.NewMockVcs()}}

	assert.Equal(t, "fallback-sha", ci.Commit())
}

func TestTeamCityCI_Configured(t *testing.T) {
	ci := &TeamCity{CIBuildName: "project"}

	assert.True(t, ci.Configured())
}
