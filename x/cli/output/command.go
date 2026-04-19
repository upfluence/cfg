package output

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/upfluence/errors"

	"github.com/upfluence/cfg"
	"github.com/upfluence/cfg/x/cli"
	"github.com/upfluence/cfg/x/cli/output/printer"
	"github.com/upfluence/cfg/x/cli/output/printer/json"
	"github.com/upfluence/cfg/x/cli/output/printer/yaml"
)

type Command[T any] interface {
	WriteSynopsis(io.Writer, cli.IntrospectionOptions) (int, error)
	WriteHelp(io.Writer, cli.IntrospectionOptions) (int, error)

	Run(context.Context, cli.CommandContext) (T, error)
}

func WrapCommand[T any](cmd Command[T], defaultPrinter printer.Printer[T], additionalPrinters ...printer.Printer[T]) cli.Command {
	printers := make(map[string]printer.Printer[T], 1+len(additionalPrinters))
	printers[defaultPrinter.Key()] = defaultPrinter

	for _, p := range additionalPrinters {
		printers[p.Key()] = p
	}

	return &wrappedCommand[T]{
		cmd:                 cmd,
		defaultOutputFormat: defaultPrinter.Key(),
		printers:            printers,
		outputConfigType:    buildOutputConfigType(printers),
	}
}

func buildOutputConfigType[T any](printers map[string]printer.Printer[T]) reflect.Type {
	keys := make([]string, 0, len(printers))

	for k := range printers {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	helpTag := fmt.Sprintf(
		"Output format (formats: [%s])",
		strings.Join(keys, " "),
	)

	return reflect.StructOf([]reflect.StructField{
		{
			Name: "OutputFormat",
			Type: reflect.TypeFor[string](),
			Tag:  reflect.StructTag(fmt.Sprintf(`flag:"o,output" help:%q`, helpTag)),
		},
	})
}

func WrapDefaultCommand[T any](cmd Command[T], additionalPrinters ...printer.Printer[T]) cli.Command {
	return WrapCommand(
		cmd,
		printer.WrapAnyPrinter[T](yaml.Printer),
		append(
			[]printer.Printer[T]{
				printer.WrapAnyPrinter[T](json.Printer),
			},
			additionalPrinters...,
		)...,
	)
}

type outputConfig struct {
	OutputFormat string `flag:"o,output" help:"Output format"`
}

type prefixedConfig struct {
	prefix string
	value  any
}

func (p *prefixedConfig) WalkPrefix() []string { return []string{"output", p.prefix} }
func (p *prefixedConfig) WalkValue() any       { return p.value }

type prefixedConfigurator struct {
	inner  cfg.Configurator
	prefix string
}

func (pc *prefixedConfigurator) Populate(ctx context.Context, in any) error {
	return pc.inner.Populate(ctx, &prefixedConfig{prefix: pc.prefix, value: in}) //nolint:wrapcheck
}

func (pc *prefixedConfigurator) WithOptions(opts ...cfg.Option) cfg.Configurator {
	return &prefixedConfigurator{
		inner:  pc.inner.WithOptions(opts...),
		prefix: pc.prefix,
	}
}

type wrappedCommand[T any] struct {
	cmd Command[T]

	defaultOutputFormat string
	printers            map[string]printer.Printer[T]
	outputConfigType    reflect.Type
}

func (wc *wrappedCommand[T]) wrapIntrospectionOptions(opts cli.IntrospectionOptions) cli.IntrospectionOptions {
	oc := reflect.New(wc.outputConfigType)
	oc.Elem().Field(0).SetString(wc.defaultOutputFormat)

	opts.Definitions = append(
		slices.Clone(opts.Definitions),
		cli.CommandDefinition{
			Configs: []any{oc.Interface()},
		},
	)

	for _, p := range wc.printers {
		def := p.CommandDefinition()
		key := p.Key()

		for i, c := range def.Configs {
			def.Configs[i] = &prefixedConfig{prefix: key, value: c}
		}

		opts.Definitions = append(opts.Definitions, def)
	}

	return opts
}

func (wc *wrappedCommand[T]) WriteSynopsis(w io.Writer, opts cli.IntrospectionOptions) (int, error) {
	return wc.cmd.WriteSynopsis(w, wc.wrapIntrospectionOptions(opts)) //nolint:wrapcheck
}

func (wc *wrappedCommand[T]) WriteHelp(w io.Writer, opts cli.IntrospectionOptions) (int, error) {
	return wc.cmd.WriteHelp(w, wc.wrapIntrospectionOptions(opts)) //nolint:wrapcheck
}

func (wc *wrappedCommand[T]) Run(ctx context.Context, cctx cli.CommandContext) error {
	var oc = outputConfig{OutputFormat: wc.defaultOutputFormat}

	if err := cctx.Configurator.Populate(ctx, &oc); err != nil {
		return errors.Wrap(err, "populate output config")
	}

	printer, ok := wc.printers[oc.OutputFormat]

	if !ok {
		return fmt.Errorf("unknown output format: %q", oc.OutputFormat)
	}

	v, err := wc.cmd.Run(ctx, cctx)

	if err != nil {
		return err //nolint:wrapcheck
	}

	cctx.Configurator = &prefixedConfigurator{
		inner:  cctx.Configurator,
		prefix: oc.OutputFormat,
	}

	return printer.Print(ctx, cctx, v) //nolint:wrapcheck
}
