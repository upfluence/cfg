package cli

import (
	"context"
	"io"
)

const defaultHelp = "no help content provided"

type helpCommand struct {
	cmd Command
}

func (hc *helpCommand) WriteHelp(w io.Writer, _ IntrospectionOptions) (int, error) {
	return io.WriteString(w, "Print this message")
}

func (hc *helpCommand) WriteSynopsis(io.Writer, IntrospectionOptions) (int, error) { return 0, nil }

func (hc *helpCommand) Run(_ context.Context, cctx CommandContext) error {
	var writeTo = writeUsage

	if hc.cmd != nil {
		writeTo = hc.cmd.WriteHelp
	} else if len(cctx.Definitions) == 0 {
		writeTo = StaticString(defaultHelp)
	}

	_, err := writeTo(cctx.Stderr, cctx.introspectionOptions())
	return err
}

type helpConfig struct {
	Help bool `flag:"h,help" help:"Display this message"`
}

func isHelpRequested(ctx context.Context, cctx CommandContext) (bool, error) {
	var c helpConfig

	if err := cctx.Configurator.Populate(ctx, &c); err != nil {
		return false, err
	}

	return c.Help, nil
}
