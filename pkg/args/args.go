package args

import (
	"bytes"
	"errors"

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
}

type VersionFlag string

const done = 1000
const unset = -1000

var Done = errors.New("version")

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
		return Done
	}
	cmd.FatalIfErrorf(err)
	return err
}
