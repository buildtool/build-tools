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

package vcs

type mockVcs struct {
	branch string
	commit string
}

// NewMockVcs returns a mockVcs with default commit and branch name
func NewMockVcs() VCS {
	return &mockVcs{
		branch: "fallback-branch",
		commit: "fallback-sha",
	}
}

// NewMockVcsWithBranch returns a mockVcs with a specific branch name
func NewMockVcsWithBranch(branch string) VCS {
	return &mockVcs{
		branch: branch,
		commit: "fallback-sha",
	}
}

func (m mockVcs) Identify(dir string) bool {
	panic("implement me")
}

func (m mockVcs) Name() string {
	panic("implement me")
}

func (m mockVcs) Branch() string {
	return m.branch
}

func (m mockVcs) Commit() string {
	return m.commit
}

var _ VCS = mockVcs{}
