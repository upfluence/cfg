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

	args map[string]string
}

func (io IntrospectionOptions) argName(arg string) string {
	if v, ok := io.args[arg]; ok {
		return v
	}

	return fmt.Sprintf("<%s>", arg)
}

func (io IntrospectionOptions) withDefinition(def CommandDefinition) IntrospectionOptions {
	return IntrospectionOptions{
		AppName:     io.AppName,
		Definitions: append(io.Definitions, def),
		Short:       io.Short,
		args:        io.args,
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

	nn, err = writeOptions(w, opts)
	n += nn

	return n, err
}

func writeOptions(w io.Writer, opts IntrospectionOptions) (int, error) {
	var cfgs []interface{}

	for _, def := range opts.Definitions {
		cfgs = append(cfgs, def.Configs...)
	}

	return help.DefaultWriter.Write(w, cfgs...)
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
			nn, err := fmt.Fprintf(w, "%s ", opts.argName(arg))
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

type EnhancedHelp struct {
	Short string
	Long  string

	Config interface{}
}

func (eh EnhancedHelp) wrapOptions(opts IntrospectionOptions) IntrospectionOptions {
	if eh.Config == nil {
		return opts
	}

	return opts.withDefinition(
		CommandDefinition{Configs: []interface{}{eh.Config}},
	)
}

func (eh EnhancedHelp) WriteHelp(w io.Writer, opts IntrospectionOptions) (int, error) {
	if opts.Short {
		return io.WriteString(w, eh.Short)
	}

	n, err := io.WriteString(w, "Description:\n\n")

	if err != nil {
		return n, err
	}

	nn, err := io.WriteString(w, eh.Long)
	n += nn

	if err != nil {
		return n, err
	}

	opts = eh.wrapOptions(opts)

	nn, err = io.WriteString(w, "\n\n")
	n += nn

	if err != nil {
		return n, err
	}

	nn, err = io.WriteString(w, "Usage:\n\t")
	n += nn

	if err != nil {
		return n, err
	}

	nn, err = writeSynopsis(w, opts)
	n += nn

	if err != nil {
		return n, err
	}

	nn, err = io.WriteString(w, "\n\n")
	n += nn

	if err != nil {
		return n, err
	}

	nn, err = writeOptions(w, opts)
	n += nn

	return n, err
}

func (eh EnhancedHelp) WriteSynopsis(w io.Writer, opts IntrospectionOptions) (int, error) {
	return writeSynopsis(w, eh.wrapOptions(opts))
}
