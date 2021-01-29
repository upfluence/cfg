package cli

import (
	"context"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

type SubCommand struct {
	Variable string

	ShortHelp IntrospectionFunc

	Commands map[string]Command
}

func (sc SubCommand) variable(defs []CommandDefinition) string {
	if sc.Variable != "" {
		return sc.Variable
	}

	var n int

	for _, def := range defs {
		n += len(def.Args)
	}

	return fmt.Sprintf("arg_%d", n+1)

}

func (sc SubCommand) definition(defs []CommandDefinition) CommandDefinition {
	return CommandDefinition{Args: []string{sc.variable(defs)}}
}

func (sc SubCommand) writeUsage(w io.Writer, opts IntrospectionOptions) (int, error) {
	var n, err = writeUsage(
		w,
		opts.withDefinition(sc.definition(opts.Definitions)),
	)

	if err != nil {
		return n, err
	}

	nn, err := io.WriteString(w, "\n")
	n += nn

	return n, err
}

func (sc SubCommand) WriteHelp(w io.Writer, opts IntrospectionOptions) (int, error) {
	var n int

	if sc.ShortHelp != nil {
		nn, err := sc.ShortHelp(w, opts)
		n += nn

		if err != nil {
			return n, err
		}
	}

	if sc.ShortHelp == nil && opts.Short {
		nn, err := sc.writeUsage(w, opts)
		n += nn

		if err != nil {
			return n, err
		}
	}

	if opts.Short {
		return 0, nil
	}

	if sc.ShortHelp != nil {
		nn, err := io.WriteString(w, "\n")
		n += nn

		if err != nil {
			return n, err
		}
	} else {
		nn, err := sc.writeUsage(w, opts)
		n += nn

		if err != nil {
			return n, err
		}
	}

	nn, err := io.WriteString(w, "Available sub commands: \n")
	n += nn

	if err != nil {
		return n, err
	}

	nn, err = sc.WriteSynopsis(w, opts)
	n += nn

	return n, err
}

func (sc SubCommand) WriteSynopsis(w io.Writer, opts IntrospectionOptions) (int, error) {
	if opts.Short {
		return fmt.Fprintf(w, "<%s>", sc.variable(opts.Definitions))
	}

	var (
		n int

		tw = tabwriter.NewWriter(w, 4, 4, 2, ' ', tabwriter.TabIndent)
		ks = make([]string, 0, len(sc.Commands))
	)

	opts = IntrospectionOptions{Short: true}

	for k := range sc.Commands {
		ks = append(ks, k)
	}

	sort.Strings(ks)

	for _, k := range ks {
		cmd := sc.Commands[k]
		nn, err := fmt.Fprintf(tw, "\t%s  \t  ", k)
		n += nn

		if err != nil {
			return n, err
		}

		nn, err = cmd.WriteHelp(tw, opts)
		n += nn

		if err != nil {
			return n, err
		}

		nn, err = io.WriteString(tw, "\n")
		n += nn

		if err != nil {
			return n, err
		}
	}

	return n, tw.Flush()
}

func (sc SubCommand) Run(ctx context.Context, cctx CommandContext) error {
	var (
		cmdKey string
		args   []string
	)

	if len(cctx.Args) > 0 {
		cmdKey = cctx.Args[0]
		args = cctx.Args[1:]
	}

	cmd, ok := sc.Commands[cmdKey]

	if !ok {
		if cmdKey == "" {
			ok, err := isHelpRequested(ctx, cctx)

			if err != nil {
				return err
			}

			if ok {
				_, err = sc.WriteHelp(cctx.Stderr, cctx.introspectionOptions())
				return err
			}
		}

		if _, err := fmt.Fprintf(
			cctx.Stderr,
			"unknown command %q, available commands:\n",
			cmdKey,
		); err != nil {
			return err
		}

		_, err := sc.WriteSynopsis(cctx.Stderr, cctx.introspectionOptions())
		return err
	}

	cctx.Definitions = append(cctx.Definitions, sc.definition(cctx.Definitions))
	cctx.Args = args

	return cmd.Run(ctx, cctx)
}
