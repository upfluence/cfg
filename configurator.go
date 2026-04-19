package cfg

import (
	"context"
	"os"
	"reflect"

	"github.com/upfluence/errors"

	"github.com/upfluence/cfg/internal/help"
	"github.com/upfluence/cfg/internal/reflectutil"
	"github.com/upfluence/cfg/internal/setter"
	"github.com/upfluence/cfg/internal/walker"
	"github.com/upfluence/cfg/provider"
	dflt "github.com/upfluence/cfg/provider/default"
	"github.com/upfluence/cfg/provider/env"
	"github.com/upfluence/cfg/provider/flags"
)

type Configurator interface {
	Populate(context.Context, interface{}) error

	WithOptions(...Option) Configurator
}

type Option func(*configurator)

func IgnoreMissingTag(c *configurator) { c.ignoreMissingTag = true }

func HonorRequired(c *configurator) { c.honorRequired = true }

func WithProviders(ps ...provider.Provider) Option {
	return func(c *configurator) { c.providers = append(c.providers, ps...) }
}

var AppendProviders = WithProviders

func OverrideProviders(ps ...provider.Provider) Option {
	return func(c *configurator) { c.providers = ps }
}

type configurator struct {
	providers        []provider.Provider
	factory          setter.Factory
	ignoreMissingTag bool
	honorRequired    bool
}

func NewDefaultConfigurator(providers ...provider.Provider) Configurator {
	cfg := newConfigurator(
		[]Option{
			HonorRequired,
			WithProviders(
				append(
					append([]provider.Provider{dflt.Provider{}}, providers...),
					env.NewDefaultProvider(),
					flags.NewDefaultProvider(),
				)...,
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

func (c *configurator) withOptions(opts []Option) *configurator {
	dup := *c

	dup.providers = append([]provider.Provider(nil), c.providers...)

	for _, opt := range opts {
		opt(&dup)
	}

	return &dup
}

func (c *configurator) WithOptions(opts ...Option) Configurator {
	return c.withOptions(opts)
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

	var set bool

	for _, p := range c.providers {
		var (
			v   string
			ok  bool
			k   string
			err error

			fqp           = provider.WrapFullyQualifiedProvider(p)
			ignoreMissing = c.ignoreMissingTag
		)

		for _, k = range walker.BuildFieldKeys(fqp, f, ignoreMissing) {
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

		set = true

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

	if !set && c.honorRequired && isRequired(f.Field) {
		return &RequiredError{Field: f.Field}
	}

	if setter.IsUnmarshaler(f.Value.Type()) {
		return walker.SkipStruct
	}

	return nil
}

func isRequired(f reflect.StructField) bool {
	v, ok := f.Tag.Lookup("required")

	if !ok {
		return false
	}

	b, err := setter.ParseBool(v)

	return err == nil && b
}
