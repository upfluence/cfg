package cli

import (
	"os"

	"github.com/upfluence/cfg"
	"github.com/upfluence/cfg/provider"
	"github.com/upfluence/cfg/provider/env"
)

type NewConfiguratorFunc func(...cfg.Option) cfg.Configurator

type Option func(*options)

func WithName(n string) Option {
	return func(o *options) { o.name = n }
}

func WithCommand(cmd Command) Option {
	return func(o *options) { o.cmd = cmd }
}

func WithConfiguratorOptions(opts ...cfg.Option) Option {
	return func(o *options) { o.opts = append(o.opts, opts...) }
}

func WithNewConfiguratorFunc(fn NewConfiguratorFunc) Option {
	return func(o *options) { o.newFunc = fn }
}

type options struct {
	name string
	args []string

	version string

	cmd     Command
	ps      []provider.Provider
	opts    []cfg.Option
	newFunc NewConfiguratorFunc
}

func defaultOptions() *options {
	return &options{
		name:    os.Args[0],
		args:    os.Args[1:],
		version: Version,
		ps:      []provider.Provider{env.NewDefaultProvider()},
		newFunc: cfg.NewConfiguratorWithOptions,
	}
}

func (o *options) command() Command {
	cmd := o.wrapCommand(o.cmd)

	if scmd, ok := cmd.(SubCommand); ok {
		if _, ok := scmd.Commands["version"]; !ok {
			scmd.Commands["version"] = o.versionCommand()
		}

		if _, ok := scmd.Commands["help"]; !ok {
			scmd.Commands["help"] = &helpCommand{cmd: cmd}
		}
	}

	return cmd
}

func (o *options) versionCommand() *versionCommand {
	return &versionCommand{name: o.name, version: o.version}
}

func (o *options) wrapCommand(cmd Command) Command {
	helpCmd := &helpCommand{cmd: cmd}

	switch tcmd := cmd.(type) {
	case nil:
		versionCmd := o.versionCommand()
		return &baseCommand{
			Command: SubCommand{
				Variable: "verb",
				Commands: map[string]Command{"version": versionCmd, "help": helpCmd},
			},
			helpCmd:    helpCmd,
			versionCmd: versionCmd,
		}
	case SubCommand:
		if tcmd.Commands == nil {
			tcmd.Commands = make(map[string]Command, 1)
		}

		if _, ok := tcmd.Commands["help"]; !ok {
			tcmd.Commands["help"] = helpCmd
		}

		for k, cmd := range tcmd.Commands {
			tcmd.Commands[k] = o.wrapCommand(cmd)
		}

		return tcmd
	case ArgumentCommand:
		tcmd.Command = o.wrapCommand(tcmd.Command)

		return tcmd
	case *baseCommand:
		return tcmd
	}

	return &baseCommand{
		Command:    cmd,
		helpCmd:    helpCmd,
		versionCmd: o.versionCommand(),
	}
}
