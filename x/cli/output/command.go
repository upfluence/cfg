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
	"github.com/upfluence/cfg/internal/walker"
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

type outputFormat struct {
	keys     []string
	selected string
}

func (of outputFormat) String() string { return of.selected }

func (of *outputFormat) Parse(s string) error {
	of.selected = s

	return nil
}

func (of outputFormat) Help() string {
	return fmt.Sprintf("Output format (formats: [%s])", strings.Join(of.keys, " "))
}

type outputConfig struct {
	OutputFormat outputFormat `flag:"o,output"`
}

func WrapCommand[T any](cmd Command[T], defaultPrinter printer.Printer[T], additionalPrinters ...printer.Printer[T]) cli.Command {
	printers := make(map[string]printer.Printer[T], 1+len(additionalPrinters))
	printers[defaultPrinter.Key()] = defaultPrinter

	for _, p := range additionalPrinters {
		printers[p.Key()] = p
	}

	keys := make([]string, 0, len(printers))

	for k := range printers {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return &wrappedCommand[T]{
		cmd:      cmd,
		printers: printers,
		outputConfig: outputConfig{
			OutputFormat: outputFormat{
				keys:     keys,
				selected: defaultPrinter.Key(),
			},
		},
	}
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

var outputAncestor = &walker.Field{
	Field: reflect.StructField{Name: "output"},
}

type prefixedConfigurator struct {
	inner  cfg.Configurator
	prefix string
}

func (pc *prefixedConfigurator) Populate(ctx context.Context, in any) error {
	return pc.inner.Populate(ctx, &walker.SubKeyPrefixed{ //nolint:wrapcheck
		Ancestor: outputAncestor,
		SubKey:   pc.prefix,
		Value:    in,
	})
}

func (pc *prefixedConfigurator) WithOptions(opts ...cfg.Option) cfg.Configurator {
	return &prefixedConfigurator{
		inner:  pc.inner.WithOptions(opts...),
		prefix: pc.prefix,
	}
}

type wrappedCommand[T any] struct {
	cmd Command[T]

	outputConfig outputConfig
	printers     map[string]printer.Printer[T]
}

func (wc *wrappedCommand[T]) wrapIntrospectionOptions(opts cli.IntrospectionOptions) cli.IntrospectionOptions {
	opts.Definitions = append(
		slices.Clone(opts.Definitions),
		cli.CommandDefinition{
			Configs: []any{&wc.outputConfig},
		},
	)

	for _, p := range wc.printers {
		def := p.CommandDefinition()
		key := p.Key()

		for i, c := range def.Configs {
			def.Configs[i] = &walker.SubKeyPrefixed{
				Ancestor: outputAncestor,
				SubKey:   key,
				Value:    c,
			}
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
	var oc = outputConfig{
		OutputFormat: outputFormat{
			keys:     wc.outputConfig.OutputFormat.keys,
			selected: wc.outputConfig.OutputFormat.selected,
		},
	}

	if err := cctx.Configurator.Populate(ctx, &oc); err != nil {
		return errors.Wrap(err, "populate output config")
	}

	p, ok := wc.printers[oc.OutputFormat.selected]

	if !ok {
		return fmt.Errorf("unknown output format: %q", oc.OutputFormat.selected)
	}

	v, err := wc.cmd.Run(ctx, cctx)

	if err != nil {
		return err //nolint:wrapcheck
	}

	cctx.Configurator = &prefixedConfigurator{
		inner:  cctx.Configurator,
		prefix: oc.OutputFormat.selected,
	}

	return p.Print(ctx, cctx, v) //nolint:wrapcheck
}
