package table

import "io"

type Formatter interface {
	WriteLine([]string) error
	Flush() error
}

type FormatterFunc func(io.Writer) Formatter
