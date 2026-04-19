package cli

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/upfluence/log"
	"github.com/upfluence/log/record"

	"github.com/upfluence/cfg"
)

type CommandContext struct {
	Configurator cfg.Configurator
	Args         []string

	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	Logger log.Logger

	Definitions []CommandDefinition

	args    map[string]string
	appName string

	env []string
	wd  string
}

func newCommandContext(a *App, cmds []string, args map[string]string, c cfg.Configurator) CommandContext {
	var wd, _ = os.Getwd()

	return CommandContext{
		Configurator: c,
		Args:         cmds,
		Stdin:        a.stdin,
		Stdout:       a.stdout,
		Stderr:       a.stderr,
		Logger:       newLogger(a.stdout, a.stderr, record.Notice),
		args:         args,
		appName:      a.name,
		wd:           wd,
		env:          os.Environ(),
	}
}

func (cctx CommandContext) SubCommand(ctx context.Context, n string, args ...string) *exec.Cmd {
	var cmd = exec.CommandContext(ctx, n, args...)

	cmd.Stdin = cctx.Stdin
	cmd.Stdout = cctx.Stdout
	cmd.Stderr = cctx.Stderr
	cmd.Env = cctx.env
	cmd.Dir = cctx.wd

	return cmd
}

func (cctx CommandContext) introspectionOptions() IntrospectionOptions {
	return IntrospectionOptions{
		AppName:     cctx.appName,
		Definitions: cctx.Definitions,
		args:        cctx.args,
	}
}
