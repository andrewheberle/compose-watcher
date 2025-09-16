package cmd

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/andrewheberle/simplecommand"
	"github.com/bep/simplecobra"
	"github.com/go-git/go-git/v6"
)

type watchCommand struct {
	// command line args
	clone    bool
	interval time.Duration
	onStart  bool

	*simplecommand.Command
}

func (c *watchCommand) Init(cd *simplecobra.Commandeer) error {
	if err := c.Command.Init(cd); err != nil {
		return err
	}

	cmd := cd.CobraCommand
	cmd.Flags().DurationVar(&c.interval, "interval", time.Minute*5, "Refresh interval")
	cmd.Flags().BoolVar(&c.onStart, "onstart", false, "Run pull and up on start")
	cmd.Flags().BoolVar(&c.clone, "clone", false, "Clone repository on start")

	return nil
}

func (c *watchCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	root, ok := cd.Root.Command.(*rootCommand)
	if !ok {
		return errors.New("problem accessing root command")
	}

	// do clone but an error is not fatal if it's because dir exists
	if c.clone {
		if err := root.gitClone(); err != nil {
			if !errors.Is(err, git.ErrTargetDirNotEmpty) {
				return err
			}
		}
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

	// do inital startup (errors are fatal at this point)
	if c.onStart {
		if err := root.composePull(); err != nil {
			return err
		}

		if err := root.composeUp(); err != nil {
			return err
		}
	}

	watchContext, cancel := context.WithCancel(ctx)
	defer cancel()

	slog.Info("starting watch", "interval", c.interval, "commit", commit.Hash)

	t := time.NewTicker(c.interval)
	for {
		select {
		case <-t.C:
			if err := root.gitPull(w); err != nil {
				slog.Error("could not pull from repository", "error", err)
				break
			}

			current, err := getCommit(r)
			if err != nil {
				slog.Error("could not get current commit of HEAD", "error", err)
				break
			}

			if commit.Hash.String() == current.Hash.String() {
				slog.Info("no changes found")
				break
			}

			commit = current
			slog.Info("changes found", "commit", commit.Hash)

			if err := root.composePull(); err != nil {
				slog.Error("could not run docker compose pull", "error", err)
				break
			}

			if err := root.composeUp(); err != nil {
				slog.Error("could not run docker compose up", "error", err)
			}

		case <-watchContext.Done():
			t.Stop()

			return watchContext.Err()
		}
	}
}
