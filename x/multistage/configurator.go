package multistage

import (
	"context"
	"fmt"

	"github.com/upfluence/cfg"
	"github.com/upfluence/cfg/provider"
)

type ProviderMode int

const (
	ProviderAppend ProviderMode = iota
	ProviderReplace
)

type ConfigurationStage[T any] struct {
	InitialConfig     T
	Mode              ProviderMode
	NextProvidersFunc func(T) ([]provider.Provider, error)
}

func (cs ConfigurationStage[T]) Next(ctx context.Context, c cfg.Configurator) ([]provider.Provider, ProviderMode, error) {
	v := cs.InitialConfig

	if err := c.Populate(ctx, &v); err != nil {
		return nil, ProviderAppend, err
	}

	ps, err := cs.NextProvidersFunc(v)

	if err != nil {
		return nil, ProviderAppend, err
	}

	return ps, cs.Mode, nil
}

type Stage interface {
	Next(context.Context, cfg.Configurator) ([]provider.Provider, ProviderMode, error)
}

type Configurator struct {
	Stages              []Stage
	InitialConfigurator cfg.Configurator
}

func (c *Configurator) initialConfigurator() cfg.Configurator {
	if c.InitialConfigurator == nil {
		return cfg.NewDefaultConfigurator()
	}

	return c.InitialConfigurator
}

func (c *Configurator) Populate(ctx context.Context, v interface{}) error {
	var tc = c.initialConfigurator()

	for _, s := range c.Stages {
		ps, m, err := s.Next(ctx, tc)

		if err != nil {
			return err
		}

		var optFunc func(...provider.Provider) cfg.Option

		switch m {
		case ProviderAppend:
			optFunc = cfg.AppendProviders
		case ProviderReplace:
			optFunc = cfg.OverrideProviders
		default:
			return fmt.Errorf("unknown ProviderMode(%+v)", m)
		}

		tc = tc.WithOptions(optFunc(ps...))
	}

	return tc.Populate(ctx, v)
}

func (c *Configurator) WithOptions(opts ...cfg.Option) cfg.Configurator {
	return &Configurator{
		Stages:              c.Stages,
		InitialConfigurator: c.initialConfigurator().WithOptions(opts...),
	}
}
