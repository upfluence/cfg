package table

import (
	"encoding/csv"
	"io"

	"github.com/upfluence/cfg/x/cli/output/printer"
)

type csvFormatter struct {
	w *csv.Writer
}

func NewCSVFormatter(w io.Writer) Formatter {
	return &csvFormatter{w: csv.NewWriter(w)}
}

func (f *csvFormatter) WriteLine(vals []string) error {
	return f.w.Write(vals) //nolint:wrapcheck
}

func (f *csvFormatter) Flush() error {
	f.w.Flush()

	return f.w.Error() //nolint:wrapcheck
}

func NewDefaultCSVPrinter[T any]() printer.Printer[[]T] {
	return NewDefaultPrinter[T]("csv", NewCSVFormatter)
}
