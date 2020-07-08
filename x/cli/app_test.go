package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	for _, tt := range []struct {
		in []string

		wantFlags []string
		wantCmds  []string
	}{
		{
			in:        []string{"-foo", "bar", "buz"},
			wantFlags: []string{"-foo", "bar"},
			wantCmds:  []string{"buz"},
		},
		{
			in:        []string{"buz", "-foo", "bar", "biz"},
			wantFlags: []string{"-foo", "bar"},
			wantCmds:  []string{"buz", "biz"},
		},
		{
			in:        []string{"buz", "-foo", "--", "biz"},
			wantFlags: []string{"-foo"},
			wantCmds:  []string{"buz", "--", "biz"},
		},
		{
			in:        []string{"buz", "-foo", "--", "biz", "-fuz"},
			wantFlags: []string{"-foo"},
			wantCmds:  []string{"buz", "--", "biz", "-fuz"},
		},
	} {
		a := &App{args: tt.in}

		cmds, flags := a.parseArgs()

		assert.Equal(t, cmds, tt.wantCmds)
		assert.Equal(t, flags, tt.wantFlags)
	}
}
