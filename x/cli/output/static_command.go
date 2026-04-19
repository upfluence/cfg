package output

import (
	"context"
	"io"

	"github.com/upfluence/errors"

	"github.com/upfluence/cfg/x/cli"
)

type StaticCommand[T any] struct {
	Help     cli.IntrospectionFunc
	Synopsis cli.IntrospectionFunc

	Execute func(context.Context, cli.CommandContext) (T, error)
}

func (sc StaticCommand[T]) WriteHelp(w io.Writer, opts cli.IntrospectionOptions) (int, error) {
	if sc.Help == nil {
		return 0, nil
	}

	return sc.Help(w, opts)
}

func (sc StaticCommand[T]) WriteSynopsis(w io.Writer, opts cli.IntrospectionOptions) (int, error) {
	if sc.Synopsis == nil {
		return 0, nil
	}

	return sc.Synopsis(w, opts)
}

func (sc StaticCommand[T]) Run(ctx context.Context, cctx cli.CommandContext) (T, error) {
	return sc.Execute(ctx, cctx)
}

type DefaultStaticCommandOption[C any, T any] func(*defaultStaticCommandOptions[C, T])

func WithShortHelp[C any, T any](h string) DefaultStaticCommandOption[C, T] {
	return func(opts *defaultStaticCommandOptions[C, T]) {
		opts.shortHelp = h
	}
}

func WithLongHelp[C any, T any](h string) DefaultStaticCommandOption[C, T] {
	return func(opts *defaultStaticCommandOptions[C, T]) {
		opts.longHelp = h
	}
}

func WithDefaultConfig[C any, T any](c C) DefaultStaticCommandOption[C, T] {
	return func(opts *defaultStaticCommandOptions[C, T]) {
		opts.defaultConfig = c
	}
}

type defaultStaticCommandOptions[C any, T any] struct {
	defaultConfig C

	shortHelp string
	longHelp  string
}

func (dsco *defaultStaticCommandOptions[C, T]) help() cli.EnhancedHelp {
	return cli.EnhancedHelp{
		Short:  dsco.shortHelp,
		Long:   dsco.longHelp,
		Config: &dsco.defaultConfig,
	}
}

func DefaultStaticCommand[C any, T any](fn func(context.Context, cli.CommandContext, C) (T, error), opts ...DefaultStaticCommandOption[C, T]) StaticCommand[T] {
	var o defaultStaticCommandOptions[C, T]

	for _, opt := range opts {
		opt(&o)
	}

	h := o.help()

	return StaticCommand[T]{
		Help:     h.WriteHelp,
		Synopsis: h.WriteSynopsis,
		Execute: func(ctx context.Context, cctx cli.CommandContext) (T, error) {
			var zero T

			config := o.defaultConfig

			if err := cctx.Configurator.Populate(ctx, &config); err != nil {
				return zero, errors.Wrap(err, "populate config")
			}

			return fn(ctx, cctx, config)
		},
	}
}
