package main

import (
	"context"
	"fmt"

	"github.com/upfluence/cfg/x/cli"
)

type config struct {
	Namespace     string `flag:"n,namespace"`
	RessourceType string `arg:"resource_type"`
	RessourceName string `arg:"resource_name"`
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

func buildObjectCommand(cmd, desc string) cli.Command {
	return cli.ArgumentCommand{
		Variable: "resource_type",
		Command: cli.ArgumentCommand{
			Variable: "resource_name",
			Command: cli.StaticCommand{
				Help:    cli.StaticString(desc),
				Execute: leafCommandExecutor(cmd).execute,
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
					"describe": buildObjectCommand("describe", "describe remote object"),
					"edit":     buildObjectCommand("edit", "edit remote object"),
				},
			},
		),
	).Run(context.Background())
}
