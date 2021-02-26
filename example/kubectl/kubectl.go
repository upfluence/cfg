package main

import (
	"context"
	"fmt"

	"github.com/upfluence/cfg/x/cli"
)

type config struct {
	Namespace     string `flag:"n,namespace" help:"namespace for this entity"`
	RessourceType string `arg:"resource_type" flag:"-" env:"-"`
	RessourceName string `arg:"resource_name" flag:"-" env:"-"`
}

type leafCommandExecutor string

func (lce leafCommandExecutor) execute(ctx context.Context, cctx cli.CommandContext) error {
	var c config

	if err := cctx.Configurator.Populate(ctx, &c); err != nil {
		return err
	}

	fmt.Fprintf(
		cctx.Stdout,
		"action=%q namespace=%q ressourceType=%q ressourceName=%q",
		lce,
		c.Namespace,
		c.RessourceType,
		c.RessourceName,
	)

	return nil
}

func buildObjectCommand(cmd string, eh cli.EnhancedHelp) cli.Command {
	return cli.ArgumentCommand{
		Variable: "resource_type",
		Command: cli.ArgumentCommand{
			Variable: "resource_name",
			Command: cli.StaticCommand{
				Help:     eh.WriteHelp,
				Synopsis: eh.WriteSynopsis,
				Execute:  leafCommandExecutor(cmd).execute,
			},
		},
	}
}

func main() {
	cli.NewApp(
		cli.WithName("kubectl"),
		cli.WithCommand(
			cli.SubCommand{
				Variable: "verb",
				Commands: map[string]cli.Command{
					"describe": buildObjectCommand(
						"describe",
						cli.EnhancedHelp{
							Short:  "describe remote object",
							Long:   "this is the long help of this command",
							Config: &config{},
						},
					),
					"edit": buildObjectCommand(
						"edit",
						cli.EnhancedHelp{
							Short:  "edit remote object",
							Long:   "this is the long help of this command",
							Config: &config{},
						},
					),
				},
			},
		),
	).Run(context.Background())
}
