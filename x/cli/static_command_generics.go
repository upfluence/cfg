//go:build go1.18

package cli

import "context"

type DefaultStaticCommandOption[T any] func(*defaultStaticCommandOptions[T])

func WithShortHelp[T any](h string) DefaultStaticCommandOption[T] {
	return func(opts *defaultStaticCommandOptions[T]) {
		opts.shortHelp = h
	}
}

func WithLongHelp[T any](h string) DefaultStaticCommandOption[T] {
	return func(opts *defaultStaticCommandOptions[T]) {
		opts.longHelp = h
	}
}

func WithDefaultConfig[T any](c T) DefaultStaticCommandOption[T] {
	return func(opts *defaultStaticCommandOptions[T]) {
		opts.defaultConfig = c
	}
}

type defaultStaticCommandOptions[T any] struct {
	defaultConfig T

	shortHelp string
	longHelp  string
}

func (dsco *defaultStaticCommandOptions[T]) help() EnhancedHelp {
	return EnhancedHelp{
		Short:  dsco.shortHelp,
		Long:   dsco.longHelp,
		Config: &dsco.defaultConfig,
	}
}

func DefaultStaticCommand[T any](fn func(context.Context, CommandContext, T) error, opts ...DefaultStaticCommandOption[T]) StaticCommand {
	var o defaultStaticCommandOptions[T]

	for _, opt := range opts {
		opt(&o)
	}

	h := o.help()

	return StaticCommand{
		Help:     h.WriteHelp,
		Synopsis: h.WriteSynopsis,
		Execute: func(ctx context.Context, cctx CommandContext) error {
			config := o.defaultConfig

			if err := cctx.Configurator.Populate(ctx, &config); err != nil {
				return err
			}

			return fn(ctx, cctx, config)
		},
	}
}
