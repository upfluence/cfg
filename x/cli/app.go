package cli

import (
	"context"
	"io"
	"os"

	"github.com/upfluence/cfg"
	"github.com/upfluence/cfg/provider"
	pflags "github.com/upfluence/cfg/provider/flags"
)

type App struct {
	ps []provider.Provider

	name string
	args []string
	cmd  Command
}

func NewApp(opts ...Option) *App {
	o := defaultOptions()

	for _, opt := range opts {
		opt(o)
	}

	return &App{
		ps:   o.ps,
		name: o.name,
		args: o.args,
		cmd:  o.command(),
	}
}

func (a *App) parseArgs() ([]string, []string) {
	var (
		cmds  []string
		flags []string

		isFlag bool
	)

	for _, arg := range a.args {
		if len(arg) == 0 {
			continue
		}

		if arg == "--" {
			isFlag = false
			cmds = append(cmds, arg)
			continue
		}

		if arg[0] == '-' {
			isFlag = true
			flags = append(flags, arg)
			continue
		}

		if isFlag {
			isFlag = false
			flags = append(flags, arg)
			continue
		}

		cmds = append(cmds, arg)
	}

	return cmds, flags
}

func (a *App) commandContext() CommandContext {
	cmds, flags := a.parseArgs()

	return CommandContext{
		Configurator: cfg.NewConfigurator(
			append([]provider.Provider{pflags.NewProvider(flags)}, a.ps...)...,
		),
		Args:   cmds,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

func (a *App) Run(ctx context.Context) {
	var code int

	if err := a.cmd.Run(ctx, a.commandContext()); err != nil {
		code = 1

		if serr, ok := err.(interface{ StatusCode() int }); ok {
			code = serr.StatusCode()
		}

		io.WriteString(os.Stderr, err.Error())
	}

	os.Stdout.Sync()
	os.Stderr.Sync()

	os.Exit(code)
}
