package cfg

import (
	"context"
	"io"
	"os"

	"github.com/upfluence/cfg/internal/help"
)

type helpConfig struct {
	Help bool `flag:"h,help" env:"HELP"`
}

type helpConfigurator struct {
	*configurator

	hw     *help.Writer
	stderr io.Writer
}

func (hc *helpConfigurator) WithOptions(opts ...Option) Configurator {
	dup := *hc
	dup.configurator = hc.configurator.withOptions(opts)

	return &dup
}

func (hc *helpConfigurator) Populate(ctx context.Context, in interface{}) error {
	var cfg helpConfig

	if err := hc.configurator.Populate(ctx, &cfg); err != nil {
		return err
	}

	if cfg.Help {
		_ = hc.PrintDefaults(in)

		os.Exit(2)
	}

	return hc.configurator.Populate(ctx, in)
}

func (hc *helpConfigurator) PrintDefaults(in interface{}) error {
	var _, err = hc.hw.Write(hc.stderr, in)
	return err
}
