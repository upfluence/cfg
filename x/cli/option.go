package cli

import (
	"os"

	"github.com/upfluence/cfg/provider"
	"github.com/upfluence/cfg/provider/env"
)

type Option func(*options)

func WithName(n string) Option {
	return func(o *options) { o.name = n }
}

func WithCommand(cmd Command) Option {
	return func(o *options) { o.cmd = cmd }
}

type options struct {
	name string
	args []string

	version string

	cmd Command
	ps  []provider.Provider
}

func defaultOptions() *options {
	return &options{
		name:    os.Args[0],
		args:    os.Args[1:],
		version: Version,
		ps:      []provider.Provider{env.NewDefaultProvider()},
	}
}

func (o *options) command() Command {
	cmd := o.cmd

	versionCmd := &versionCommand{name: o.name, version: o.version}
	helpCmd := &helpCommand{cmd: cmd}

	if cmd == nil {
		cmd = SubCommand{}
	}

	if scmd, ok := cmd.(SubCommand); ok {
		if _, ok := scmd["version"]; !ok {
			scmd["version"] = versionCmd
		}

		if _, ok := scmd["help"]; !ok {
			scmd["help"] = helpCmd
		}
	}

	return &baseCommand{
		Command:    cmd,
		helpCmd:    helpCmd,
		versionCmd: versionCmd,
	}
}
