package cli

import (
	"context"
	"fmt"
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

func (vc *versionCommand) Run(ctx context.Context, cctx CommandContext) error {
	fmt.Fprintf(cctx.Stdout, "%s/%s\n", vc.name, vc.version)
	return nil
}
