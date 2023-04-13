package cfg

import (
	"context"
	"os"

	"github.com/upfluence/errors"

	"github.com/upfluence/cfg/internal/help"
	"github.com/upfluence/cfg/internal/reflectutil"
	"github.com/upfluence/cfg/internal/setter"
	"github.com/upfluence/cfg/internal/walker"
	"github.com/upfluence/cfg/provider"
	"github.com/upfluence/cfg/provider/env"
	"github.com/upfluence/cfg/provider/flags"
)

type Configurator interface {
	Populate(context.Context, interface{}) error
}

type Option func(*configurator)

func IgnoreMissingTag(c *configurator) { c.ignoreMissingTag = true }

func WithProviders(ps ...provider.Provider) Option {
	return func(c *configurator) { c.providers = append(c.providers, ps...) }
}

type configurator struct {
	providers        []provider.Provider
	factory          setter.Factory
	ignoreMissingTag bool
}

func NewDefaultConfigurator(providers ...provider.Provider) Configurator {
	cfg := newConfigurator(
		[]Option{
			WithProviders(
				append(providers, env.NewDefaultProvider(), flags.NewDefaultProvider())...,
			),
		},
	)

	return &helpConfigurator{
		configurator: cfg,
		hw:           &help.Writer{Providers: cfg.providers, Factory: cfg.factory},
		stderr:       os.Stderr,
	}
}

func NewConfigurator(providers ...provider.Provider) Configurator {
	return NewConfiguratorWithOptions(WithProviders(providers...))
}

func NewConfiguratorWithOptions(opts ...Option) Configurator {
	return newConfigurator(opts)
}

func newConfigurator(opts []Option) *configurator {
	var c = configurator{factory: setter.DefaultFactory}

	for _, opt := range opts {
		opt(&c)
	}

	return &c
}

func (c *configurator) Populate(ctx context.Context, in interface{}) error {
	return walker.Walk(
		in,
		func(f *walker.Field) error { return c.walkFunc(ctx, f) },
	)
}

func (c *configurator) walkFunc(ctx context.Context, f *walker.Field) error {
	s := c.factory.Build(f.Field.Type)

	if s == nil {
		return nil
	}

	for _, p := range c.providers {
		var (
			v   string
			ok  bool
			k   string
			err error
		)

		for _, k = range walker.BuildFieldKeys(p.StructTag(), f, c.ignoreMissingTag) {
			v, ok, err = p.Provide(ctx, k)

			if err != nil {
				return errors.WithStack(
					&ProvidingError{
						Err:      err,
						Key:      k,
						Field:    f.Field,
						Provider: p,
					},
				)
			}

			if ok {
				break
			}
		}

		if !ok {
			continue
		}

		if err := s.Set(
			v,
			reflectutil.IndirectedValue(f.Value).FieldByName(f.Field.Name),
		); err != nil {
			return errors.WithStack(
				&SettingError{
					Err:      err,
					Key:      k,
					Value:    v,
					Field:    f.Field,
					Provider: p,
				},
			)
		}
	}

	if setter.IsUnmarshaler(f.Value.Type()) {
		return walker.SkipStruct
	}

	return nil
}
