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
	"path"
	"path/filepath"
	"strings"
)

type Github struct {
	*Common
	CICommit     string `env:"GITHUB_SHA"`
	CIBuildName  string `env:"RUNNER_WORKSPACE"`
	CIRepository string `env:"GITHUB_REPOSITORY"`
	CIBranchName string `env:"GITHUB_REF"`
}

var _ CI = &Github{}

func (c *Github) Name() string {
	return "Github"
}

func (c *Github) BranchReplaceSlash() string {
	return branchReplaceSlash(c.Branch())
}

func (c *Github) BuildName() string {
	// GITHUB_REPOSITORY is "<owner>/<repository>" and names the repository directly.
	// Prefer it over deriving the name from RUNNER_WORKSPACE, whose layout differs per
	// platform: GitHub Actions uses /home/runner/work/<repository>, while Gitea Actions
	// uses /workspace/<owner>/<repository> and exports its parent, so the last path
	// element there is the owner rather than the repository.
	if c.CIRepository != "" {
		return c.Common.BuildName(path.Base(c.CIRepository))
	}
	if c.CIBuildName != "" {
		return c.Common.BuildName(filepath.Base(c.CIBuildName))
	}
	return c.Common.BuildName(c.CIBuildName)
}

func (c *Github) Branch() string {
	if strings.HasPrefix(c.CIBranchName, "refs/heads") {
		return c.Common.Branch(strings.TrimPrefix(c.CIBranchName, "refs/heads/"))
	} else if strings.HasPrefix(c.CIBranchName, "refs/tags") {
		return c.Common.Branch(strings.TrimPrefix(c.CIBranchName, "refs/tags/"))
	} else {
		return c.Common.Branch(c.CIBranchName)
	}
}

func (c *Github) Commit() string {
	return c.Common.Commit(c.CICommit)
}

func (c *Github) Configured() bool {
	return c.CICommit != ""
}
