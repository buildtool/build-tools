package args

import (
	"errors"
	"fmt"
	"io"

	"github.com/alecthomas/kong"

	"github.com/buildtool/build-tools/pkg/config"
	"github.com/buildtool/build-tools/pkg/version"
)

type Globals struct {
	Version VersionFlag `name:"version" help:"Print args information and quit"`
	Verbose bool        `short:"v" help:"Enable verbose mode"`
	Config  ConfigFlag  ``
}

type VersionFlag string

const done = 1000
const unset = -1000

var Done = errors.New("version")

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	_, _ = fmt.Fprintf(app.Stdout, vars["version"])
	app.Exit(done)
	return nil
}

type ConfigFlag string

func (v ConfigFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v ConfigFlag) IsBool() bool                         { return true }
func (v ConfigFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	cfg, err := config.Load(vars["dir"], app.Stdout)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(app.Stdout, "Current config\n")
	_ = cfg.Print(app.Stdout)
	app.Exit(done)
	return nil
}

func ParseArgs(dir string, out, eout io.Writer, osArgs []string, info version.Info, variables interface{}) error {
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
		kong.UsageOnError(),
		kong.Description(info.Description),
		kong.Writers(out, eout),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{"version": info.String(), "dir": dir},
	)
	_, err := cmd.Parse(osArgs)
	if exitCode == done {
		return Done
	}
	cmd.FatalIfErrorf(err)
	return err
}
