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

//go:build !prod
// +build !prod

package config

import (
	"os"

	git2 "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func InitRepo(dir string) *git2.Repository {
	repo, _ := git2.PlainInit(dir, false)
	return repo
}

func InitRepoWithCommit(dir string) (plumbing.Hash, *git2.Repository) {
	repo := InitRepo(dir)
	tree, _ := repo.Worktree()
	file, _ := os.CreateTemp(dir, "file")
	_, _ = file.WriteString("test")
	_ = file.Close()
	_, _ = tree.Add(file.Name())
	hash, _ := tree.Commit("Test", &git2.CommitOptions{Author: &object.Signature{Email: "test@example.com"}})

	return hash, repo
}
