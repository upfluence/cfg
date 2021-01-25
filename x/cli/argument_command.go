package cli

import (
	"context"
	"fmt"
	"io"
)

type ArgumentCommand struct {
	Variable string
	Command  Command
}

func (sc ArgumentCommand) WriteSynopsis(w io.Writer) (int, error) {
	return sc.Command.WriteSynopsis(w)
}

func (sc ArgumentCommand) WriteHelp(w io.Writer) (int, error) {
	return sc.Command.WriteHelp(w)
}

func (sc ArgumentCommand) Run(ctx context.Context, cctx CommandContext) error {
	if len(cctx.Args) == 0 {
		if _, err := fmt.Fprintf(
			cctx.Stderr,
			"no argument found for variable %q, follow the synopsis:\n",
			sc.Variable,
		); err != nil {
			return err
		}

		_, err := sc.WriteSynopsis(cctx.Stderr)
		return err
	}

	subject := cctx.Args[0]
	cctx.Args = cctx.Args[1:]
	cctx.args[sc.Variable] = subject

	return sc.Command.Run(ctx, cctx)
}
