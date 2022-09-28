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

type TeamCity struct {
	*Common
	CICommit     string `env:"BUILD_VCS_NUMBER"`
	CIBuildName  string `env:"TEAMCITY_PROJECT_NAME"`
	CIBranchName string `env:"BUILD_VCS_BRANCH"`
}

var _ CI = &TeamCity{}

func (c TeamCity) Name() string {
	return "TeamCity"
}

func (c TeamCity) BranchReplaceSlash() string {
	return branchReplaceSlash(c.Branch())
}

func (c TeamCity) BuildName() string {
	return c.Common.BuildName(c.CIBuildName)
}

func (c TeamCity) Branch() string {
	return c.Common.Branch(c.CIBranchName)
}

func (c TeamCity) Commit() string {
	return c.Common.Commit(c.CICommit)
}

func (c TeamCity) Configured() bool {
	return c.CIBuildName != ""
}
