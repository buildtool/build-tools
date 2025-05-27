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

package args

import (
	"bytes"
	"errors"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/apex/log"

	"github.com/buildtool/build-tools/pkg/cli"
	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/version"
)

type Globals struct {
	Version VersionFlag `name:"version" help:"Print args information and exit"`
	Verbose VerboseFlag `short:"v" help:"Enable verbose mode"`
	Config  ConfigFlag  `help:"Print parsed config and exit"`
	StdIn   io.Reader   `kong:"-"`
}

type VersionFlag string

const (
	done  = 1000
	unset = -1000
)

var ErrDone = errors.New("version")

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	log.Info(vars["version"])
	app.Exit(done)
	return nil
}

type ConfigFlag string

func (v ConfigFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v ConfigFlag) IsBool() bool                         { return true }
func (v ConfigFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	cfg, err := config.Load(vars["dir"])
	if err != nil {
		return err
	}
	out := &bytes.Buffer{}
	_ = cfg.Print(out)
	log.Infof("Current config\n%s", out.String())
	app.Exit(done)
	return nil
}

type VerboseFlag string

func (v VerboseFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VerboseFlag) IsBool() bool                         { return true }
func (v VerboseFlag) BeforeApply(*kong.Kong, kong.Vars) error {
	log.SetLevel(log.DebugLevel)
	return nil
}

func (v *Globals) AfterApply() error {
	v.StdIn = os.Stdin
	return nil
}

func ParseArgs(dir string, osArgs []string, info version.Info, variables interface{}) error {
	exitCode := unset
	cmd, _ := kong.New(
		variables,
		kong.Name(info.Name),
		kong.Exit(func(i int) {
			if exitCode == unset {
				exitCode = i
			}
		}),
		kong.Help(func(options kong.HelpOptions, ctx *kong.Context) error {
			_ = kong.DefaultHelpPrinter(options, ctx)
			ctx.Exit(done)
			return nil
		}),
		kong.Description(info.Description),
		kong.Writers(cli.NewWriter(log.Log), cli.NewWriter(log.Log)),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{"version": info.String(), "dir": dir},
	)
	_, err := cmd.Parse(osArgs)
	if exitCode == done {
		return ErrDone
	}
	cmd.FatalIfErrorf(err)
	return err
}
