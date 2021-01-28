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

func (sc ArgumentCommand) definition() CommandDefinition {
	return CommandDefinition{Args: []string{sc.Variable}}
}

func (sc ArgumentCommand) WriteSynopsis(w io.Writer, opts IntrospectionOptions) (int, error) {
	return sc.Command.WriteSynopsis(w, opts.withDefinition(sc.definition()))
}

func (sc ArgumentCommand) WriteHelp(w io.Writer, opts IntrospectionOptions) (int, error) {
	return sc.Command.WriteHelp(w, opts.withDefinition(sc.definition()))
}

func (sc ArgumentCommand) Run(ctx context.Context, cctx CommandContext) error {
	if len(cctx.Args) == 0 {
		ok, err := isHelpRequested(ctx, cctx)

		if err != nil {
			return err
		}

		if ok {
			_, err := sc.WriteHelp(
				cctx.Stderr,
				IntrospectionOptions{Definitions: cctx.Definitions},
			)

			return err
		}

		if _, err := fmt.Fprintf(
			cctx.Stderr,
			"no argument found for variable %q, follow the synopsis:\n",
			sc.Variable,
		); err != nil {
			return err
		}

		_, err = sc.WriteSynopsis(
			cctx.Stderr,
			IntrospectionOptions{Definitions: cctx.Definitions},
		)

		return err
	}

	subject := cctx.Args[0]
	cctx.Args = cctx.Args[1:]
	cctx.Definitions = append(cctx.Definitions, sc.definition())
	cctx.args[sc.Variable] = subject

	return sc.Command.Run(ctx, cctx)
}
