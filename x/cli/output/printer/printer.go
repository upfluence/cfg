package printer

import (
	"context"

	"github.com/upfluence/cfg/x/cli"
)

type Printer[T any] interface {
	Key() string
	CommandDefinition() cli.CommandDefinition
	Print(context.Context, cli.CommandContext, T) error
}

type AnyPrinter = Printer[any]

func WrapAnyPrinter[T any](ap AnyPrinter) Printer[T] {
	return &wrappedAnyPrinter[T]{AnyPrinter: ap}
}

type wrappedAnyPrinter[T any] struct {
	AnyPrinter
}

func (wap *wrappedAnyPrinter[T]) Print(ctx context.Context, cctx cli.CommandContext, v T) error {
	return wap.AnyPrinter.Print(ctx, cctx, v) //nolint:wrapcheck
}
