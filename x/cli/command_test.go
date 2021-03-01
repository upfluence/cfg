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

type mockConfig1 struct {
	Example string `flag:"e,example"`
}

type mockConfig2 struct {
	Other string `flag:"o,other"`
}

type mockCommand struct {
	Command
}

func (mc *mockCommand) WriteHelp(w io.Writer, opts IntrospectionOptions) (int, error) {
	return mc.Command.WriteHelp(
		w,
		opts.withDefinition(
			CommandDefinition{Configs: []interface{}{&mockConfig1{}}},
		),
	)
}

func TestRun(t *testing.T) {
	staticCmd := StaticCommand{
		Help:     StaticString("help foo"),
		Synopsis: SynopsisWriter(&mockConfig1{}),
		Execute: func(_ context.Context, cctx CommandContext) error {
			_, err := io.WriteString(cctx.Stdout, "success")
			return err
		},
	}

	eh := EnhancedHelp{
		Short:  "enhanced short help",
		Long:   "enhanced long help",
		Config: &mockConfig2{},
	}

	subCmd := SubCommand{Commands: map[string]Command{"foo": staticCmd}}
	nestedCmd := SubCommand{
		Commands: map[string]Command{
			"foo": subCmd,
			"bar": &mockCommand{
				Command: StaticCommand{
					Help:     HelpWriter(&mockConfig2{}),
					Synopsis: SynopsisWriter(&mockConfig2{}),
					Execute: func(_ context.Context, cctx CommandContext) error {
						_, err := io.WriteString(cctx.Stdout, "success")
						return err
					},
				},
			},
			"buz": &mockCommand{
				Command: StaticCommand{
					Help:     eh.WriteHelp,
					Synopsis: eh.WriteSynopsis,
					Execute: func(_ context.Context, cctx CommandContext) error {
						_, err := io.WriteString(cctx.Stdout, "success")
						return err
					},
				},
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
			wantErr: defaultHelp,
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
			opts: []Option{WithCommand(argCmd)},
			args: []string{"-y"},
			wantErr: `no argument found for variable "buz", follow the synopsis:
cli-test <buz> `,
		},
		{
			opts:    []Option{WithCommand(argCmd)},
			args:    []string{"-h"},
			wantErr: "usage: cli-test <buz> ",
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
foo help foo
help Print this message
version Print the app version `,
		},
		{
			args: []string{"-h"},
			opts: []Option{WithCommand(subCmd)},
			wantErr: `usage: cli-test <arg_1>
Available sub commands:
foo help foo
help Print this message
version Print the app version `,
		},
		{
			args: []string{"-h"},
			opts: []Option{WithCommand(nestedCmd)},
			wantErr: `usage: cli-test <arg_1>
Available sub commands:
bar usage: [-e, --example] [-o, --other]
buz enhanced short help
foo usage: <arg_1>
help Print this message
version Print the app version `,
		},
		{
			args: []string{"foo", "-h"},
			opts: []Option{WithCommand(nestedCmd)},
			wantErr: `usage: cli-test <arg_1> <arg_2>
Available sub commands:
foo help foo
help Print this message
version Print the app version `,
		},
		{
			args: []string{"bar", "-h"},
			opts: []Option{WithCommand(nestedCmd)},
			wantErr: `usage: cli-test <arg_1> [-e, --example] [-o, --other]
Arguments:
- Example: string (env: EXAMPLE, flag: -e, --example)
- Other: string (env: OTHER, flag: -o, --other) `,
		},
		{
			args: []string{"buz", "-h"},
			opts: []Option{WithCommand(nestedCmd)},
			wantErr: `Description:
enhanced long help
Usage:
cli-test <arg_1> [-e, --example] [-o, --other]
Arguments:
- Example: string (env: EXAMPLE, flag: -e, --example)
- Other: string (env: OTHER, flag: -o, --other) `,
		},
		{
			args:    []string{"foo", "foo", "-h"},
			opts:    []Option{WithCommand(nestedCmd)},
			wantErr: `usage: cli-test <arg_1> <arg_2> help foo`,
		},
	} {
		var (
			outBuf bytes.Buffer
			errBuf bytes.Buffer

			a = NewApp(append([]Option{WithName("cli-test")}, tt.opts...)...)
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
