package cmd

import (
	"context"
	"errors"
	"log/slog"

	"github.com/andrewheberle/simplecommand"
	"github.com/bep/simplecobra"
	"github.com/go-git/go-git/v6"
)

type checkCommand struct {
	// command line args
	force bool

	*simplecommand.Command
}

func (c *checkCommand) Init(cd *simplecobra.Commandeer) error {
	if err := c.Command.Init(cd); err != nil {
		return err
	}

	cmd := cd.CobraCommand
	cmd.Flags().BoolVar(&c.force, "force", false, "Always start containers via Docker Compose")

	return nil
}

func (c *checkCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	root, ok := cd.Root.Command.(*rootCommand)
	if !ok {
		return errors.New("problem accessing root command")
	}

	r, err := git.PlainOpen(root.directory)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	commit, err := getCommit(r)
	if err != nil {
		return err
	}

	// do initial pull
	if err := root.gitPull(w); err != nil {
		return err
	}

	current, err := getCommit(r)
	if err != nil {
		return err
	}

	// do compose pull/up if forced or changes
	if c.force || commit.Hash.String() != current.Hash.String() {
		if err := root.composePull(); err != nil {
			return err
		}

		if err := root.composeUp(); err != nil {
			return err
		}
	} else {
		slog.Info("no changes detected and not forced")
	}

	return nil
}
