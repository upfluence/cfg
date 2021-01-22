package cli

import (
	"context"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
)

type SubCommand map[string]Command

func (sc SubCommand) WriteHelp(w io.Writer) (int, error) {
	n, err := io.WriteString(w, "Sub commands available: \n")

	if err != nil {
		return n, err
	}

	nn, err := sc.WriteSynopsis(w)

	return n + nn, err
}

func (sc SubCommand) WriteSynopsis(w io.Writer) (int, error) {
	var n int

	tw := tabwriter.NewWriter(w, 4, 4, 2, ' ', tabwriter.TabIndent)

	ks := make([]string, 0, len(sc))

	for k := range sc {
		ks = append(ks, k)
	}

	sort.Strings(ks)

	for _, k := range ks {
		cmd := sc[k]
		nn, err := fmt.Fprintf(tw, "\t%s  \t  ", k)

		n += nn

		if err != nil {
			return n, err
		}

		nn, err = cmd.WriteHelp(tw)

		n += nn

		if err != nil {
			return n, err
		}

		nn, err = io.WriteString(tw, "  ")

		n += nn

		if err != nil {
			return n, err
		}

		nn, err = cmd.WriteSynopsis(tw)

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

	cmd, ok := sc[cmdKey]

	if !ok {
		if _, err := fmt.Fprintf(
			cctx.Stderr,
			"unknown command %q, available commands:\n",
			cmdKey,
		); err != nil {
			return err
		}

		_, err := sc.WriteSynopsis(cctx.Stderr)
		return err
	}

	cctx.Args = args

	return cmd.Run(ctx, cctx)
}
