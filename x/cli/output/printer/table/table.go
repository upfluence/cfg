package table

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/upfluence/cfg/x/cli/output/printer"
)

type tabwriterFormatter struct {
	tw *tabwriter.Writer
}

func NewTabwriterFormatter(w io.Writer) Formatter {
	return &tabwriterFormatter{
		tw: tabwriter.NewWriter(w, 0, 0, 2, ' ', 0),
	}
}

func (f *tabwriterFormatter) WriteLine(vals []string) error {
	_, err := fmt.Fprintln(f.tw, strings.Join(vals, "\t"))

	return err //nolint:wrapcheck
}

func (f *tabwriterFormatter) Flush() error {
	return f.tw.Flush() //nolint:wrapcheck
}

func NewDefaultTablePrinter[T any]() printer.Printer[[]T] {
	return NewDefaultPrinter[T]("table", NewTabwriterFormatter)
}
