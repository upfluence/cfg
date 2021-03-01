package cli

import (
	"context"
	"fmt"
	"io"
)

type argProvider map[string]string

func (ap argProvider) StructTag() string { return "arg" }
func (ap argProvider) Provide(_ context.Context, k string) (string, bool, error) {
	v, ok := ap[k]
	return v, ok, nil
}

type Command interface {
	WriteSynopsis(io.Writer, IntrospectionOptions) (int, error)
	WriteHelp(io.Writer, IntrospectionOptions) (int, error)

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

	if bc.Command == nil {
		_, err := fmt.Fprintf(cctx.Stderr, "command not implemented")
		return err
	}

	return bc.Command.Run(ctx, cctx)
}
