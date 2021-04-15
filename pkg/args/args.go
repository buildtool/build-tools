package args

import (
	"errors"
	"fmt"
	"io"

	"github.com/alecthomas/kong"

	"github.com/buildtool/build-tools/pkg/version"
)

type Globals struct {
	Version VersionFlag `name:"version" help:"Print args information and quit"`
	Verbose bool        `short:"v" help:"Enable verbose mode"`
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

func ParseArgs(out, eout io.Writer, osArgs []string, info version.Info, variables interface{}) error {
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
		kong.Vars{"version": info.String()},
	)
	_, err := cmd.Parse(osArgs)
	if exitCode == done {
		return Done
	}
	cmd.FatalIfErrorf(err)
	return err
}
