package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	staticCmd := StaticCommand{
		Help:     StaticString("help foo"),
		Synopsis: StaticString("foo synopsis"),
		Execute: func(_ context.Context, cctx CommandContext) error {
			_, err := io.WriteString(cctx.Stdout, "success")
			return err
		},
	}

	subCmd := SubCommand{
		"foo": StaticCommand{
			Help:     StaticString("help foo"),
			Synopsis: StaticString("foo synopsis"),
			Execute: func(_ context.Context, cctx CommandContext) error {
				_, err := io.WriteString(cctx.Stdout, "success")
				return err
			},
		},
	}

	argCmd := ArgumentCommand{
		Variable: "buz",
		Command: StaticCommand{
			Execute: func(ctx context.Context, cctx CommandContext) error {
				var c = struct {
					Buz string `arg:"buz"`
				}{}

				if err := cctx.Configurator.Populate(ctx, &c); err != nil {
					return err
				}

				_, err := fmt.Fprintf(cctx.Stdout, "<%s>", c.Buz)
				return err
			},
		},
	}

	for _, tt := range []struct {
		opts []Option
		args []string

		wantOut string
		wantErr string

		err error
	}{
		{
			wantErr: `unknown command "", available commands:
help         Print this message
version      Print the app version
`,
		},
		{
			args: []string{"subcommand"},
			wantErr: `unknown command "subcommand", available commands:
help         Print this message
version      Print the app version
`,
		},
		{
			args:    []string{"-h"},
			wantOut: defaultHelp,
		},
		{
			args:    []string{"-v"},
			opts:    []Option{WithName("testapp")},
			wantOut: "testapp/dirty\n",
		},
		{
			opts:    []Option{WithCommand(staticCmd)},
			wantOut: "success",
		},
		{
			opts:    []Option{WithCommand(argCmd)},
			args:    []string{"foo", "-y"},
			wantOut: "<foo>",
		},
		{
			opts:    []Option{WithCommand(argCmd)},
			args:    []string{"-y"},
			wantErr: "no argument found for variable \"buz\", follow the synopsys: ",
		},
		{
			args:    []string{"foo"},
			opts:    []Option{WithCommand(subCmd)},
			wantOut: "success",
		},
		{
			args: []string{"buz"},
			opts: []Option{WithCommand(subCmd)},
			wantErr: `unknown command "buz", available commands:
foo help foo foo synopsis
help Print this message
version Print the app version `,
		},
		{
			args: []string{"-h"},
			opts: []Option{WithCommand(subCmd)},
			wantOut: `Sub commands available:
foo help foo foo synopsis
help Print this message
version Print the app version `,
		},
	} {
		var (
			outBuf bytes.Buffer
			errBuf bytes.Buffer

			a = NewApp(tt.opts...)
		)

		a.args = tt.args

		cctx := a.commandContext()
		cctx.Stdout = &outBuf
		cctx.Stderr = &errBuf

		err := a.cmd.Run(context.Background(), cctx)

		assert.Equal(t, canonicalString(outBuf.String()), canonicalString(tt.wantOut))
		assert.Equal(t, canonicalString(errBuf.String()), canonicalString(tt.wantErr))
		assert.Equal(t, err, tt.err)
	}
}

func canonicalString(v string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(v, " ")
}
