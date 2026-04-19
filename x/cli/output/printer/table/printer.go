package table

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/upfluence/errors"

	"github.com/upfluence/cfg/internal/walker"
	"github.com/upfluence/cfg/provider"
	"github.com/upfluence/cfg/x/cli"
	"github.com/upfluence/cfg/x/cli/output/printer"
)

const key = "table"

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
	columns      []string
	extractValue func(T, string) string
}

func NewPrinter[T any](cols []string, extractValue func(T, string) string) printer.Printer[[]T] {
	return &tablePrinter[T]{
		columns:      cols,
		extractValue: extractValue,
	}
}

func NewDefaultPrinter[T any]() printer.Printer[[]T] {
	var (
		cols          []string
		indexByColumn = make(map[string][]int)
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

			col, ok := buildColumnName(f)

			if !ok {
				return nil
			}

			cols = append(cols, col)
			indexByColumn[col] = buildIndex(f)

			return nil
		},
	)

	return &tablePrinter[T]{
		columns: cols,
		extractValue: func(v T, col string) string {
			rv := reflect.ValueOf(v)
			idx, ok := indexByColumn[col]

			if !ok {
				return ""
			}

			return fmt.Sprintf("%v", rv.FieldByIndex(idx).Interface())
		},
	}
}

func (p *tablePrinter[T]) Key() string { return key }

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

var columnProvider = provider.WrapFullyQualifiedProvider(
	provider.NewStaticProvider("table", nil, nil),
)

func buildColumnName(f *walker.Field) (string, bool) {
	keys := walker.BuildFieldKeys(columnProvider, f, false)

	if len(keys) == 0 {
		return "", false
	}

	return keys[0], true
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
	tw := tabwriter.NewWriter(cctx.Stdout, 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintln(tw, strings.Join(cols, "\t")); err != nil {
		return err //nolint:wrapcheck
	}

	for _, v := range vs {
		vals := make([]string, len(cols))

		for i, col := range cols {
			vals[i] = p.extractValue(v, col)
		}

		if _, err := fmt.Fprintln(tw, strings.Join(vals, "\t")); err != nil {
			return err //nolint:wrapcheck
		}
	}

	return tw.Flush() //nolint:wrapcheck
}
