package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/upfluence/log/record"
)

func TestLogger(t *testing.T) {
	for _, tc := range []struct {
		name       string
		haveArgs   []string
		wantStdout string
		wantStderr string
	}{
		{
			name:       "default filters debug and writes notice to stdout",
			wantStdout: "notice\n",
			wantStderr: "err\n",
		},
		{
			name:       "verbose writes debug to stdout",
			haveArgs:   []string{"--verbose"},
			wantStdout: "debug\nnotice\n",
			wantStderr: "err\n",
		},
		{
			name:       "log-level debug writes debug to stdout",
			haveArgs:   []string{"--log-level", "debug"},
			wantStdout: "debug\nnotice\n",
			wantStderr: "err\n",
		},
		{
			name:       "log-level error filters notice and writes error to stderr",
			haveArgs:   []string{"--log-level", "error"},
			wantStderr: "err\n",
		},
		{
			name:       "log-level overrides verbose",
			haveArgs:   []string{"--verbose", "--log-level", "error"},
			wantStderr: "err\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				stdout, stderr bytes.Buffer

				a = NewApp(
					WithName("test"),
					WithCommand(StaticCommand{
						Execute: func(_ context.Context, cctx CommandContext) error {
							cctx.Logger.Debug("debug")
							cctx.Logger.Notice("notice")
							cctx.Logger.Error("err")

							return nil
						},
					}),
				)
			)

			a.args = tc.haveArgs

			cctx := a.commandContext()
			cctx.Stdout = &stdout
			cctx.Stderr = &stderr

			err := a.cmd.Run(context.Background(), cctx)

			require.NoError(t, err)
			assert.Equal(t, tc.wantStdout, stdout.String())
			assert.Equal(t, tc.wantStderr, stderr.String())
		})
	}
}

func TestLogLevelParse(t *testing.T) {
	for _, tc := range []struct {
		name    string
		have    string
		want    logLevel
		wantErr string
	}{
		{
			name: "debug",
			have: "debug",
			want: logLevel(record.Debug),
		},
		{
			name: "error",
			have: "error",
			want: logLevel(record.Error),
		},
		{
			name:    "unknown",
			have:    "bogus",
			wantErr: `unknown log level "bogus"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var ll logLevel

			err := ll.Parse(tc.have)

			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, ll)
		})
	}
}
