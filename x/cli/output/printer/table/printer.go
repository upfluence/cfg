package table

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"text/tabwriter"

	"github.com/upfluence/cfg/internal/walker"
	"github.com/upfluence/cfg/x/cli"
	"github.com/upfluence/cfg/x/cli/output/printer"
)

const key = "table"

type tablePrinter[T any] struct {
	columns      []string
	extractValue func(T, string) string
}

func NewPrinter[T any](columns []string, extractValue func(T, string) string) printer.Printer[[]T] {
	return &tablePrinter[T]{
		columns:      columns,
		extractValue: extractValue,
	}
}

func NewDefaultPrinter[T any]() printer.Printer[[]T] {
	var (
		columns       []string
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

			col := buildColumnName(f)
			columns = append(columns, col)
			indexByColumn[col] = buildIndex(f)

			return nil
		},
	)

	return &tablePrinter[T]{
		columns: columns,
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
	return cli.CommandDefinition{}
}

func buildColumnName(f *walker.Field) string {
	var parts []string

	for a := f.Ancestor; a != nil; a = a.Ancestor {
		parts = append(parts, strings.ToUpper(a.Field.Name))
	}

	// reverse to get root-first order
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	parts = append(parts, strings.ToUpper(f.Field.Name))

	return strings.Join(parts, ".")
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

func (p *tablePrinter[T]) Print(_ context.Context, cctx cli.CommandContext, vs []T) error {
	tw := tabwriter.NewWriter(cctx.Stdout, 0, 0, 2, ' ', 0)

	if _, err := fmt.Fprintln(tw, strings.Join(p.columns, "\t")); err != nil {
		return err //nolint:wrapcheck
	}

	for _, v := range vs {
		vals := make([]string, len(p.columns))

		for i, col := range p.columns {
			vals[i] = p.extractValue(v, col)
		}

		if _, err := fmt.Fprintln(tw, strings.Join(vals, "\t")); err != nil {
			return err //nolint:wrapcheck
		}
	}

	return tw.Flush() //nolint:wrapcheck
}
