package table

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/upfluence/errors"

	"github.com/upfluence/cfg/internal/walker"
	"github.com/upfluence/cfg/provider"
	"github.com/upfluence/cfg/x/cli"
	"github.com/upfluence/cfg/x/cli/output/printer"
)

type columns struct {
	available []string
	selected  []string
}

func (c columns) String() string { return strings.Join(c.selected, ",") }

func (c *columns) Parse(s string) error {
	c.selected = strings.Split(s, ",")

	return nil
}

func (c columns) Help() string {
	return fmt.Sprintf("Columns to display (available: [%s])", strings.Join(c.available, " "))
}

type config struct {
	Columns columns `flag:"columns"`
}

type tablePrinter[T any] struct {
	key          string
	columns      []string
	extractValue func(T, string) string
	formatter    FormatterFunc
}

func NewPrinter[T any](key string, ff FormatterFunc, cols []string, extractValue func(T, string) string) printer.Printer[[]T] {
	return &tablePrinter[T]{
		key:          key,
		columns:      cols,
		extractValue: extractValue,
		formatter:    ff,
	}
}

func introspectType[T any](key string) ([]string, func(T, string) string) {
	var (
		cols          []string
		indexByColumn = make(map[string][]int)

		colProvider = provider.WrapFullyQualifiedProvider(
			provider.NewStaticProvider(key, nil, nil),
		)
	)

	walker.Walk( //nolint:errcheck
		reflect.New(reflect.TypeFor[T]()).Interface(),
		func(f *walker.Field) error {
			ft := f.Field.Type

			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}

			if ft.Kind() == reflect.Struct {
				return nil
			}

			keys := walker.BuildFieldKeys(colProvider, f, false)

			if len(keys) == 0 {
				return nil
			}

			col := keys[0]
			cols = append(cols, col)
			indexByColumn[col] = buildIndex(f)

			return nil
		},
	)

	return cols, func(v T, col string) string {
		rv := reflect.ValueOf(v)
		idx, ok := indexByColumn[col]

		if !ok {
			return ""
		}

		return fmt.Sprintf("%v", rv.FieldByIndex(idx).Interface())
	}
}

func NewDefaultPrinter[T any](key string, ff FormatterFunc) printer.Printer[[]T] {
	cols, extractValue := introspectType[T](key)

	return NewPrinter[T](key, ff, cols, extractValue)
}

func (p *tablePrinter[T]) Key() string { return p.key }

func (p *tablePrinter[T]) CommandDefinition() cli.CommandDefinition {
	return cli.CommandDefinition{
		Configs: []any{
			&config{
				Columns: columns{
					available: p.columns,
					selected:  p.columns,
				},
			},
		},
	}
}

func buildIndex(f *walker.Field) []int {
	var idx []int

	for a := f.Ancestor; a != nil; a = a.Ancestor {
		idx = append(idx, a.Field.Index...)
	}

	// reverse ancestor indices to get root-first order
	for i, j := 0, len(idx)-1; i < j; i, j = i+1, j-1 {
		idx[i], idx[j] = idx[j], idx[i]
	}

	return append(idx, f.Field.Index...)
}

func (p *tablePrinter[T]) Print(ctx context.Context, cctx cli.CommandContext, vs []T) error {
	var cfg = config{
		Columns: columns{
			available: p.columns,
			selected:  p.columns,
		},
	}

	if err := cctx.Configurator.Populate(ctx, &cfg); err != nil {
		return errors.Wrap(err, "populate table config")
	}

	cols := cfg.Columns.selected
	f := p.formatter(cctx.Stdout)

	if err := f.WriteLine(cols); err != nil {
		return err //nolint:wrapcheck
	}

	for _, v := range vs {
		vals := make([]string, len(cols))

		for i, col := range cols {
			vals[i] = p.extractValue(v, col)
		}

		if err := f.WriteLine(vals); err != nil {
			return err //nolint:wrapcheck
		}
	}

	return f.Flush()
}
