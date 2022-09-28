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

package main

import (
	"os"

	"github.com/apex/log"

	"github.com/buildtool/build-tools/pkg/args"
	"github.com/buildtool/build-tools/pkg/build"
	"github.com/buildtool/build-tools/pkg/cli"
	ver "github.com/buildtool/build-tools/pkg/version"
)

var (
	version              = "dev"
	commit               = "none"
	date                 = "unknown"
	exitFunc             = os.Exit
	handler  log.Handler = cli.New(os.Stdout)
)

func main() {
	log.SetHandler(handler)
	dir, _ := os.Getwd()

	var buildArgs build.Args
	info := ver.Info{
		Name:        "build",
		Description: "performs a docker build and tags the resulting image",
		Version:     version,
		Commit:      commit,
		Date:        date,
	}
	err := args.ParseArgs(dir, os.Args[1:], info, &buildArgs)
	if err != nil {
		if err != args.Done {
			exitFunc(-1)
			return
		} else {
			exitFunc(0)
			return
		}
	}

	if err := build.DoBuild(dir, buildArgs); err != nil {
		log.Error(err.Error())
		exitFunc(-1)
		return
	}
	exitFunc(0)
}
