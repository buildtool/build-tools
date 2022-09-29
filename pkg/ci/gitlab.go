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

type Gitlab struct {
	*Common
	CICommit     string `env:"CI_COMMIT_SHA"`
	CIBuildName  string `env:"CI_PROJECT_NAME"`
	CIBranchName string `env:"CI_COMMIT_REF_NAME"`
}

var _ CI = &Gitlab{}

func (c *Gitlab) Name() string {
	return "Gitlab"
}

func (c *Gitlab) BranchReplaceSlash() string {
	return branchReplaceSlash(c.Branch())
}

func (c *Gitlab) BuildName() string {
	return c.Common.BuildName(c.CIBuildName)
}

func (c *Gitlab) Branch() string {
	return c.Common.Branch(c.CIBranchName)
}

func (c *Gitlab) Commit() string {
	return c.Common.Commit(c.CICommit)
}

func (c *Gitlab) Configured() bool {
	return c.CIBuildName != ""
}
