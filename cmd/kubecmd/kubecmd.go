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
	"fmt"
	"io"
	"os"

	"github.com/apex/log"

	"github.com/buildtool/build-tools/pkg/cli"
	"github.com/buildtool/build-tools/pkg/kubecmd"
	ver "github.com/buildtool/build-tools/pkg/version"
)

var (
	version             = "dev"
	commit              = "none"
	date                = "unknown"
	out     io.Writer   = os.Stdout
	handler log.Handler = cli.New(os.Stderr)
)

func main() {
	log.SetHandler(handler)
	dir, _ := os.Getwd()
	if cmd := kubecmd.Kubecmd(dir, ver.Info{
		Name:        "kubecmd",
		Description: "Generates a kubectl command, using the configuration from .buildtools.yaml if found",
		Version:     version,
		Commit:      commit,
		Date:        date,
	}, os.Args[1:]...); cmd != nil {
		_, _ = fmt.Fprintf(out, *cmd)
	}
}
