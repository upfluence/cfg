package cli

import (
	"context"
	"io"
)

type StaticCommand struct {
	Help     IntrospectionFunc
	Synopsis IntrospectionFunc

	Execute func(context.Context, CommandContext) error
}

func (sc StaticCommand) WriteHelp(w io.Writer, opts IntrospectionOptions) (int, error) {
	if sc.Help == nil {
		return writeUsage(w, opts)
	}

	return sc.Help(w, opts)
}

func (sc StaticCommand) WriteSynopsis(w io.Writer, opts IntrospectionOptions) (int, error) {
	if sc.Synopsis == nil {
		return writeSynopsis(w, opts)
	}

	return sc.Synopsis(w, opts)
}

func (sc StaticCommand) Run(ctx context.Context, cctx CommandContext) error {
	return sc.Execute(ctx, cctx)
}
