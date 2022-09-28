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
	"strings"

	"github.com/apex/log"
	"github.com/buildtool/build-tools/pkg/vcs"
)

type CI interface {
	Name() string
	// BuildName returns the name of the current build in lowercase
	BuildName() string
	Branch() string
	BranchReplaceSlash() string
	Commit() string
	SetVCS(vcs vcs.VCS)
	SetImageName(imageName string)
	Configured() bool
}

type Common struct {
	VCS       vcs.VCS
	ImageName string
}

func (c *Common) SetVCS(vcs vcs.VCS) {
	c.VCS = vcs
}

func (c *Common) SetImageName(imageName string) {
	c.ImageName = imageName
}

func (c *Common) BuildName(name string) string {
	if c.ImageName != "" {
		log.Infof("Using %s as BuildName\n", c.ImageName)
		return c.ImageName
	}
	if name != "" {
		return strings.ToLower(name)
	}
	dir, _ := os.Getwd()
	return strings.ToLower(filepath.Base(dir))
}

func (c *Common) Branch(name string) string {
	if name != "" {
		return name
	}
	return c.VCS.Branch()
}

func (c *Common) Commit(name string) string {
	if name != "" {
		return name
	}
	return c.VCS.Commit()
}

func branchReplaceSlash(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(name, "/", "_"), " ", "_")
}

func IsValid(c CI) bool {
	return len(c.Commit()) != 0 || len(c.Branch()) != 0
}
