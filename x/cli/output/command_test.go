package output_test

import (
	"bytes"
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/upfluence/cfg/x/cli"
	"github.com/upfluence/cfg/x/cli/output"
	"github.com/upfluence/cfg/x/cli/output/printer"
	pjson "github.com/upfluence/cfg/x/cli/output/printer/json"
	pyaml "github.com/upfluence/cfg/x/cli/output/printer/yaml"
)

type testConfig struct {
	Foo string `flag:"foo,f"`
	Bar string `flag:"bar,b"`
}

type testResult struct {
	Message string `json:"message" yaml:"message"`
	Foo     string `json:"foo"     yaml:"foo"`
}

func defaultStaticCommand() output.StaticCommand[testResult] {
	return output.DefaultStaticCommand(
		func(_ context.Context, _ cli.CommandContext, c testConfig) (testResult, error) {
			return testResult{Message: "ok", Foo: c.Foo}, nil
		},
	)
}

func canonicalString(v string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(v, " ")
}

func TestRun(t *testing.T) {
	for _, tc := range []struct {
		name     string
		haveArgs []string
		haveCmd  cli.Command
		wantCode int
		wantOut  string
		wantErr  string
	}{
		{
			name:     "help shows output flag and json-indent",
			haveArgs: []string{"-h"},
			haveCmd: output.WrapDefaultCommand[testResult](
				output.DefaultStaticCommand(
					func(_ context.Context, _ cli.CommandContext, _ testConfig) (testResult, error) {
						return testResult{}, nil
					},
					output.WithShortHelp[testConfig, testResult]("test command"),
				),
			),
			wantErr: `Description:

Usage:
test-app [-o, --output] [--output.json.indent] [--foo, -f] [--bar, -b]
Arguments:
- OutputFormat: output.outputFormat Output format (formats: [json yaml]) (default: yaml) (env: OUTPUTFORMAT, flag: -o, --output)
- output.json.Indent: bool Indent JSON output (env: OUTPUT_JSON_INDENT, flag: --output.json.indent)
- Foo: string (env: FOO, flag: --foo, -f)
- Bar: string (env: BAR, flag: --bar, -b) `,
		},
		{
			name:     "help with custom printers",
			haveArgs: []string{"-h"},
			haveCmd: output.WrapCommand[testResult](
				output.DefaultStaticCommand(
					func(_ context.Context, _ cli.CommandContext, _ testConfig) (testResult, error) {
						return testResult{}, nil
					},
				),
				printer.WrapAnyPrinter[testResult](pjson.Printer),
			),
			wantErr: `Description:

Usage:
test-app [-o, --output] [--output.json.indent] [--foo, -f] [--bar, -b]
Arguments:
- OutputFormat: output.outputFormat Output format (formats: [json]) (default: json) (env: OUTPUTFORMAT, flag: -o, --output)
- output.json.Indent: bool Indent JSON output (env: OUTPUT_JSON_INDENT, flag: --output.json.indent)
- Foo: string (env: FOO, flag: --foo, -f)
- Bar: string (env: BAR, flag: --bar, -b) `,
		},
		{
			name:     "default format is yaml",
			haveArgs: []string{"--foo", "hello"},
			haveCmd:  output.WrapDefaultCommand[testResult](defaultStaticCommand()),
			wantOut:  "message: ok\nfoo: hello\n",
		},
		{
			name:     "json format",
			haveArgs: []string{"--foo", "hello", "-o", "json"},
			haveCmd:  output.WrapDefaultCommand[testResult](defaultStaticCommand()),
			wantOut:  "{\"message\":\"ok\",\"foo\":\"hello\"}\n",
		},
		{
			name:     "json format with indent",
			haveArgs: []string{"--foo", "bar", "-o", "json", "--output.json.indent"},
			haveCmd:  output.WrapDefaultCommand[testResult](defaultStaticCommand()),
			wantOut:  "{\n  \"message\": \"ok\",\n  \"foo\": \"bar\"\n}\n",
		},
		{
			name:     "explicit yaml format",
			haveArgs: []string{"--foo", "baz", "-o", "yaml"},
			haveCmd:  output.WrapDefaultCommand[testResult](defaultStaticCommand()),
			wantOut:  "message: ok\nfoo: baz\n",
		},
		{
			name:     "unknown format returns error",
			haveArgs: []string{"-o", "xml"},
			haveCmd: output.WrapDefaultCommand[testResult](
				output.DefaultStaticCommand(
					func(_ context.Context, _ cli.CommandContext, _ testConfig) (testResult, error) {
						return testResult{}, nil
					},
				),
			),
			wantCode: 1,
			wantErr:  `unknown output format: "xml"`,
		},
		{
			name:     "WrapCommand with single printer",
			haveArgs: []string{"--foo", "val"},
			haveCmd: output.WrapCommand[testResult](
				defaultStaticCommand(),
				printer.WrapAnyPrinter[testResult](pyaml.Printer),
			),
			wantOut: "message: ok\nfoo: val\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var outBuf, errBuf bytes.Buffer

			a := cli.NewApp(
				cli.WithName("test-app"),
				cli.WithCommand(tc.haveCmd),
				cli.WithArgs(tc.haveArgs),
				cli.WithStdout(&outBuf),
				cli.WithStderr(&errBuf),
			)

			msg, code := a.Execute(context.Background())

			assert.Equal(t, tc.wantCode, code)
			assert.Equal(t, tc.wantOut, outBuf.String())
			assert.Equal(t, canonicalString(tc.wantErr), canonicalString(errBuf.String()+msg))
		})
	}
}
