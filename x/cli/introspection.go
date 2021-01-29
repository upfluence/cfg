package cli

import (
	"fmt"
	"io"

	"github.com/upfluence/cfg/internal/help"
	"github.com/upfluence/cfg/internal/synopsis"
)

type CommandDefinition struct {
	Args    []string
	Configs []interface{}
}

type IntrospectionOptions struct {
	AppName     string
	Definitions []CommandDefinition
	Short       bool
}

func (io IntrospectionOptions) withDefinition(def CommandDefinition) IntrospectionOptions {
	return IntrospectionOptions{
		AppName:     io.AppName,
		Definitions: append(io.Definitions, def),
		Short:       io.Short,
	}
}

type IntrospectionFunc func(io.Writer, IntrospectionOptions) (int, error)

func StaticString(v string) IntrospectionFunc {
	return func(w io.Writer, opts IntrospectionOptions) (int, error) {
		var n int

		if !opts.Short {
			nn, err := writeUsage(w, opts)
			n += nn

			if err != nil {
				return n, err
			}

			if nn > 0 {
				nn, err = io.WriteString(w, "\n")
				n += nn

				if err != nil {
					return n, err
				}
			}
		}

		nn, err := io.WriteString(w, v)
		n += nn

		return n, err
	}
}

func HelpWriter(in interface{}) IntrospectionFunc {
	return func(w io.Writer, opts IntrospectionOptions) (int, error) {
		return writeHelp(
			w,
			opts.withDefinition(CommandDefinition{Configs: []interface{}{in}}),
		)
	}
}

func writeHelp(w io.Writer, opts IntrospectionOptions) (int, error) {
	var n int

	if len(opts.Definitions) == 0 {
		return 0, nil
	}

	nn, err := io.WriteString(w, "usage: ")
	n += nn

	if err != nil {
		return n, err
	}

	nn, err = writeSynopsis(w, opts)
	n += nn

	if err != nil || opts.Short {
		return n, err
	}

	nn, err = io.WriteString(w, "\n")
	n += nn

	if err != nil {
		return n, err
	}

	var cfgs []interface{}

	for _, def := range opts.Definitions {
		cfgs = append(cfgs, def.Configs...)
	}

	nn, err = help.DefaultWriter.Write(w, cfgs...)
	n += nn

	return n, err
}

func SynopsisWriter(in interface{}) IntrospectionFunc {
	return func(w io.Writer, opts IntrospectionOptions) (int, error) {
		return writeSynopsis(
			w,
			opts.withDefinition(CommandDefinition{Configs: []interface{}{in}}),
		)
	}
}

func writeSynopsis(w io.Writer, opts IntrospectionOptions) (int, error) {
	var n int

	if opts.AppName != "" {
		nn, err := fmt.Fprintf(w, "%s ", opts.AppName)
		n += nn

		if err != nil {
			return n, err
		}
	}

	for _, def := range opts.Definitions {
		for _, arg := range def.Args {
			nn, err := fmt.Fprintf(w, "<%s> ", arg)
			n += nn

			if err != nil {
				return n, err
			}
		}

		for _, cfg := range def.Configs {
			nn, err := synopsis.DefaultWriter.Write(w, cfg)
			n += nn

			if err != nil {
				return n, err
			}
		}
	}

	return n, nil
}

func writeUsage(w io.Writer, opts IntrospectionOptions) (int, error) {
	opts.Short = true

	return writeHelp(w, opts)
}
