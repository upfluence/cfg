package yaml

import (
	"context"

	"gopkg.in/yaml.v3"

	"github.com/upfluence/cfg/x/cli"
	"github.com/upfluence/cfg/x/cli/output/printer"
)

const key = "yaml"

var Printer printer.AnyPrinter = anyPrinter{}

type anyPrinter struct{}

func (anyPrinter) Key() string { return key }

func (anyPrinter) CommandDefinition() cli.CommandDefinition {
	return cli.CommandDefinition{}
}

func (anyPrinter) Print(_ context.Context, cctx cli.CommandContext, v any) error {
	return yaml.NewEncoder(cctx.Stdout).Encode(v) //nolint:wrapcheck
}
