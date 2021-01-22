package cli

import (
	"context"
	"io"
)

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
