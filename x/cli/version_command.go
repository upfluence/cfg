package cli

import (
	"context"
	"io"
)

var Version = "dirty"

type versionCommand struct {
	name    string
	version string
}

func (vc *versionCommand) WriteSynopsis(io.Writer, IntrospectionOptions) (int, error) { return 0, nil }

func (vc *versionCommand) WriteHelp(w io.Writer, _ IntrospectionOptions) (int, error) {
	return io.WriteString(w, "Print the app version")
}

func (vc *versionCommand) Run(_ context.Context, cctx CommandContext) error {
	cctx.Logger.Noticef("%s/%s", vc.name, vc.version)

	return nil
}
