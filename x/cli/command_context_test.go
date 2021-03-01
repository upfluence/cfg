package cli

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubCommand(t *testing.T) {
	os.Setenv("FOO", "bar")

	var (
		buf bytes.Buffer

		cctx = newCommandContext("", nil, nil, nil)
	)

	cctx.Stdout = &buf
	cctx.Stderr = &buf

	err := cctx.SubCommand(
		context.Background(),
		"/bin/bash",
		"-c",
		`echo "$FOO"`,
	).Run()

	assert.NoError(t, err)
	assert.Equal(t, "bar\n", buf.String())
}
