package cli

import (
	"context"
	"io"
)

type StaticCommand struct {
	Help     func(io.Writer) (int, error)
	Synopsis func(io.Writer) (int, error)

	Execute func(context.Context, CommandContext) error
}

func (sc StaticCommand) WriteHelp(w io.Writer) (int, error) {
	if sc.Help == nil {
		return io.WriteString(w, "no help provided")
	}

	return sc.Help(w)
}

func (sc StaticCommand) WriteSynopsis(w io.Writer) (int, error) {
	if sc.Synopsis == nil {
		return 0, nil
	}

	return sc.Synopsis(w)
}

func (sc StaticCommand) Run(ctx context.Context, cctx CommandContext) error {
	return sc.Execute(ctx, cctx)
}
