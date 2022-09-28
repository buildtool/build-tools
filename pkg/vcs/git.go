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

import (
	"github.com/apex/log"
	git2 "github.com/go-git/go-git/v5"
)

type git struct {
	CommonVCS
	repo *git2.Repository
}

func (v *git) Identify(dir string) bool {
	options := &git2.PlainOpenOptions{DetectDotGit: true}
	repo, err := git2.PlainOpenWithOptions(dir, options)
	if err != nil {
		return false
	}
	v.repo = repo
	ref, err := repo.Head()
	if err != nil {
		log.Debugf("Unable to fetch head: %s\n", err)
		return false
	}
	v.CurrentCommit = ref.Hash().String()
	v.CurrentBranch = ref.Name().Short()

	return true
}

func (v *git) Name() string {
	return "Git"
}

var _ VCS = &git{}
