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

package vcs

// VCS represent the VersionControlSystem used
type VCS interface {
	// Identify returns true if it is the expected VCS type (based on information found in dir)
	Identify(dir string) bool
	// Name returns the name of the VCS
	Name() string
	// Branch returns the current branch
	Branch() string
	// Commit returns the current commit
	Commit() string
}

// CommonVCS contains functions shared by all VCSs
type CommonVCS struct {
	CurrentBranch string
	CurrentCommit string
}

// Branch returns the current branch
func (v CommonVCS) Branch() string {
	return v.CurrentBranch
}

// Commit returns the current commit
func (v CommonVCS) Commit() string {
	return v.CurrentCommit
}

var systems = []VCS{&git{}}

// Identify tries to identify the actual VCS
func Identify(dir string) VCS {
	for _, vcs := range systems {
		if vcs.Identify(dir) {
			return vcs
		}
	}

	no := &no{CommonVCS: CommonVCS{}}
	return no
}
