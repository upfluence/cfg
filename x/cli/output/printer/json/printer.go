package json

import (
	"context"
	"encoding/json"

	"github.com/upfluence/errors"

	"github.com/upfluence/cfg/x/cli"
	"github.com/upfluence/cfg/x/cli/output/printer"
)

const key = "json"

var Printer printer.AnyPrinter = anyPrinter{}

type config struct {
	Indent bool `flag:"indent" help:"Indent JSON output"`
}

type anyPrinter struct{}

func (anyPrinter) Key() string { return key }

func (anyPrinter) CommandDefinition() cli.CommandDefinition {
	return cli.CommandDefinition{
		Configs: []any{&config{}},
	}
}

func (anyPrinter) Print(ctx context.Context, cctx cli.CommandContext, v any) error {
	var cfg config

	if err := cctx.Configurator.Populate(ctx, &cfg); err != nil {
		return errors.Wrap(err, "populate json config")
	}

	enc := json.NewEncoder(cctx.Stdout)

	if cfg.Indent {
		enc.SetIndent("", "  ")
	}

	return enc.Encode(v) //nolint:wrapcheck
}
