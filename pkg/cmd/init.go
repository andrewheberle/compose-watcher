package cmd

import (
	"context"
	"errors"

	"github.com/andrewheberle/simplecommand"
	"github.com/bep/simplecobra"
)

type initCommand struct {
	*simplecommand.Command
}

func (c *initCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	root, ok := cd.Root.Command.(*rootCommand)
	if !ok {
		return errors.New("problem accessing root command")
	}

	return root.gitClone()
}
