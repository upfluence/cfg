package cli

import (
	"context"
	"io"

	"github.com/upfluence/log/record"
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
	Help     bool      `flag:"h,help"    help:"Display this message"`
	Version  bool      `flag:"v,version" help:"Display the app version"`
	Verbose  bool      `flag:"verbose"   help:"Enable verbose logging"`
	LogLevel *logLevel `flag:"log-level" help:"Set the log level (debug, info, notice, warning, error)"`
}

func (bc baseConfig) logLevel() record.Level {
	if bc.LogLevel != nil {
		return record.Level(*bc.LogLevel)
	}

	if bc.Verbose {
		return record.Debug
	}

	return record.Notice
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

	cctx.Logger = newLogger(cctx.Stdout, cctx.Stderr, cfg.logLevel())

	if cfg.Version {
		return bc.versionCmd.Run(ctx, cctx)
	}

	if cfg.Help {
		return bc.helpCmd.Run(ctx, cctx)
	}

	if bc.Command == nil {
		cctx.Logger.Error("command not implemented")

		return nil
	}

	return bc.Command.Run(ctx, cctx)
}
