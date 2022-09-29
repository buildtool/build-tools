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

type Azure struct {
	*Common
	CICommit     string `env:"BUILD_SOURCEVERSION"`
	CIBuildName  string `env:"BUILD_REPOSITORY_NAME"`
	CIBranchName string `env:"BUILD_SOURCEBRANCHNAME"`
}

var _ CI = &Azure{}

func (c Azure) Name() string {
	return "Azure"
}

func (c Azure) BranchReplaceSlash() string {
	return branchReplaceSlash(c.Branch())
}

func (c Azure) BuildName() string {
	return c.Common.BuildName(c.CIBuildName)
}

func (c Azure) Branch() string {
	return c.Common.Branch(c.CIBranchName)
}

func (c Azure) Commit() string {
	return c.Common.Commit(c.CICommit)
}

func (c Azure) Configured() bool {
	return c.CIBuildName != ""
}
