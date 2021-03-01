package cli

import (
	"context"
	"io"
	"os"
	"os/exec"

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
		nested bool
	)

	for _, arg := range a.args {
		if len(arg) == 0 {
			continue
		}

		if nested {
			cmds = append(cmds, arg)
			continue
		}

		if arg == "--" {
			isFlag = false
			cmds = append(cmds, arg)
			nested = true
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
	var (
		cmds, flags = a.parseArgs()
		args        = make(map[string]string)
		ps          = append(
			[]provider.Provider{pflags.NewProvider(flags), argProvider(args)},
			a.ps...,
		)
	)

	return newCommandContext(a.name, cmds, args, cfg.NewConfigurator(ps...))
}

func (a *App) Run(ctx context.Context) {
	var code int

	if err := a.cmd.Run(ctx, a.commandContext()); err != nil {
		code = 1

		switch serr := err.(type) {
		case *exec.ExitError:
			os.Exit(serr.ExitCode())
		case interface{ StatusCode() int }:
			code = serr.StatusCode()
		}

		_, _ = io.WriteString(os.Stderr, err.Error()+"\n")
	}

	os.Stdout.Sync()
	os.Stderr.Sync()

	os.Exit(code)
}
