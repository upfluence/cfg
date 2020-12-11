package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"

	"github.com/upfluence/cfg"
	"github.com/upfluence/cfg/internal/help"
	"github.com/upfluence/cfg/internal/walker"
	"github.com/upfluence/cfg/provider/flags"
)

type CommandContext struct {
	Configurator cfg.Configurator
	Args         []string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type Command interface {
	WriteSynopsis(io.Writer) (int, error)
	WriteHelp(io.Writer) (int, error)

	Run(context.Context, CommandContext) error
}

type baseConfig struct {
	Help    bool `flag:"h,help" help:"Display this message"`
	Version bool `flag:"v,version" help:"Display the app version"`
}

type baseCommand struct {
	Command

	helpCmd    Command
	versionCmd Command
}

func (bc *baseCommand) Run(ctx context.Context, cctx CommandContext) error {
	var cfg baseConfig

	if err := cctx.Configurator.Populate(ctx, &cfg); err != nil {
		return err
	}

	if cfg.Version {
		return bc.versionCmd.Run(ctx, cctx)
	}

	if cfg.Help {
		return bc.helpCmd.Run(ctx, cctx)
	}

	return bc.Command.Run(ctx, cctx)
}

var Version = "dirty"

type versionCommand struct {
	name    string
	version string
}

func (vc *versionCommand) WriteSynopsis(io.Writer) (int, error) { return 0, nil }

func (vc *versionCommand) WriteHelp(w io.Writer) (int, error) {
	return io.WriteString(w, "Print the app version")
}

func (vc *versionCommand) Run(ctx context.Context, cctx CommandContext) error {
	fmt.Fprintf(cctx.Stdout, "%s/%s\n", vc.name, vc.version)
	return nil
}

const defaultHelp = "no help content provided"

type helpCommand struct {
	cmd Command
}

func (hc *helpCommand) WriteHelp(w io.Writer) (int, error) {
	return io.WriteString(w, "Print this message")
}

func (hc *helpCommand) WriteSynopsis(io.Writer) (int, error) { return 0, nil }

func (hc *helpCommand) Run(_ context.Context, cctx CommandContext) error {
	var writeTo = func(w io.Writer) (int, error) {
		return io.WriteString(w, defaultHelp)
	}

	if hc.cmd != nil {
		writeTo = hc.cmd.WriteHelp
	}

	_, err := writeTo(cctx.Stdout)

	return err
}

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

func StaticString(v string) func(io.Writer) (int, error) {
	return func(w io.Writer) (int, error) { return io.WriteString(w, v) }
}

type StaticCommand struct {
	Help     func(io.Writer) (int, error)
	Synopsis func(io.Writer) (int, error)

	Execute func(context.Context, CommandContext) error
}

func (sc StaticCommand) WriteHelp(w io.Writer) (int, error) {
	if sc.Help == nil {
		return io.WriteString(w, "no help provided")
	}

	return sc.Help(w)
}

func (sc StaticCommand) WriteSynopsis(w io.Writer) (int, error) {
	if sc.Synopsis == nil {
		return 0, nil
	}

	return sc.Synopsis(w)
}

func (sc StaticCommand) Run(ctx context.Context, cctx CommandContext) error {
	return sc.Execute(ctx, cctx)
}

func HelpWriter(in interface{}) func(io.Writer) (int, error) {
	return func(w io.Writer) (int, error) {
		return help.DefaultWriter.Write(w, in)
	}
}

func SynopsisWriter(in interface{}) func(io.Writer) (int, error) {
	return func(w io.Writer) (int, error) {
		var b bytes.Buffer

		if err := walker.Walk(
			in,
			func(f *walker.Field) error {
				fks := walker.BuildFieldKeys(flags.StructTag, f)

				if len(fks) == 0 {
					return nil
				}

				b.WriteRune('[')

				for i, fk := range fks {
					b.WriteRune('-')

					if len(fk) > 1 {
						b.WriteRune('-')
					}

					b.WriteString(fk)

					if i < len(fks)-1 {
						b.WriteString(", ")
					}
				}

				b.WriteString("] ")
				return nil
			},
		); err != nil {
			return 0, err
		}

		return w.Write(b.Bytes())
	}
}
