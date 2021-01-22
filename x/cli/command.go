package cli

import (
	"context"
	"io"

	"github.com/upfluence/cfg"
	"github.com/upfluence/cfg/internal/help"
	"github.com/upfluence/cfg/internal/synopsis"
)

type CommandContext struct {
	Configurator cfg.Configurator
	Args         []string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type Command interface {
	WriteSynopsis(io.Writer) (int, error)
	WriteHelp(io.Writer) (int, error)

	Run(context.Context, CommandContext) error
}

type baseConfig struct {
	Help    bool `flag:"h,help" help:"Display this message"`
	Version bool `flag:"v,version" help:"Display the app version"`
}

type baseCommand struct {
	Command

	helpCmd    Command
	versionCmd Command
}

func (bc *baseCommand) Run(ctx context.Context, cctx CommandContext) error {
	var cfg baseConfig

	if err := cctx.Configurator.Populate(ctx, &cfg); err != nil {
		return err
	}

	if cfg.Version {
		return bc.versionCmd.Run(ctx, cctx)
	}

	if cfg.Help {
		return bc.helpCmd.Run(ctx, cctx)
	}

	return bc.Command.Run(ctx, cctx)
}

func StaticString(v string) func(io.Writer) (int, error) {
	return func(w io.Writer) (int, error) { return io.WriteString(w, v) }
}

func HelpWriter(in interface{}) func(io.Writer) (int, error) {
	return func(w io.Writer) (int, error) {
		return help.DefaultWriter.Write(w, in)
	}
}

func SynopsisWriter(in interface{}) func(io.Writer) (int, error) {
	return func(w io.Writer) (int, error) {
		return synopsis.DefaultWriter.Write(w, in)
	}
}
