package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/andrewheberle/redacted-string"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
)

func (c *rootCommand) cloneOptions() (*git.CloneOptions, error) {
	opts := &git.CloneOptions{
		URL: c.url,
	}

	if c.key != "" {
		publicKeys, err := ssh.NewPublicKeysFromFile("git", c.key, c.password)
		if err != nil {
			return nil, err
		}

		opts.Auth = publicKeys
	} else if c.username != "" && c.password != "" {
		opts.Auth = &http.BasicAuth{
			Username: c.username,
			Password: c.password,
		}
	}

	return opts, nil
}

func (c *rootCommand) pullOptions() (*git.PullOptions, error) {
	opts := &git.PullOptions{
		RemoteName: "origin",
	}

	if c.key != "" {
		publicKeys, err := ssh.NewPublicKeysFromFile("git", c.key, c.password)
		if err != nil {
			return nil, err
		}

		opts.Auth = publicKeys
	} else if c.username != "" && c.password != "" {
		opts.Auth = &http.BasicAuth{
			Username: c.username,
			Password: c.password,
		}
	}

	return opts, nil
}

func getCommit(r *git.Repository) (*object.Commit, error) {
	ref, err := r.Head()
	if err != nil {
		return nil, err
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}

	return commit, nil
}

func (c *rootCommand) gitClone() error {
	opts, err := c.cloneOptions()
	if err != nil {
		return err
	}

	r, err := git.PlainClone(c.directory, opts)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	branchRef := plumbing.NewBranchReferenceName(c.branch)
	checkoutOpts := &git.CheckoutOptions{
		Branch: branchRef,
		Force:  true,
	}

	// do local checkout
	if err := w.Checkout(checkoutOpts); err != nil {
		// try to check out remote branch of same name
		remote, err := r.Remote("origin")
		if err != nil {
			return err
		}

		if err := remote.Fetch(&git.FetchOptions{
			RefSpecs: []config.RefSpec{config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", c.branch, c.branch))},
		}); err != nil {
			if err != git.NoErrAlreadyUpToDate {
				return err
			}
		}

		// try local checkout again
		if err := w.Checkout(checkoutOpts); err != nil {
			return err
		}
	}

	return nil
}

func (c *rootCommand) gitPull(w *git.Worktree) error {
	// get pull options
	opts, err := c.pullOptions()
	if err != nil {
		return err
	}

	if err := w.Pull(opts); err != nil {
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return err
		}
	}

	return nil
}

func (c *rootCommand) composePull() error {
	stdOut := new(bytes.Buffer)
	stdErr := new(bytes.Buffer)
	cmd := exec.Command("docker", "compose", "pull")
	cmd.Dir = c.directory
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	if err := cmd.Run(); err != nil {
		attrs := make([]slog.Attr, 0)
		if stdOut.Len() > 0 {
			attrs = append(attrs, slog.String("stdout", stdOut.String()))
		}
		if stdErr.Len() > 0 {
			attrs = append(attrs, slog.String("stderr", stdErr.String()))
		}
		slog.LogAttrs(context.Background(), slog.LevelError, "docker compose pull", attrs...)
		return fmt.Errorf("error during docker compose pull: %w", err)
	}

	return nil
}

func (c *rootCommand) composeUp() error {
	stdOut := new(bytes.Buffer)
	stdErr := new(bytes.Buffer)
	cmd := exec.Command("docker", "compose", "up", "-d")
	cmd.Dir = c.directory
	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	if err := cmd.Run(); err != nil {
		attrs := make([]slog.Attr, 0)
		if stdOut.Len() > 0 {
			attrs = append(attrs, slog.String("stdout", stdOut.String()))
		}
		if stdErr.Len() > 0 {
			attrs = append(attrs, slog.String("stderr", stdErr.String()))
		}
		slog.LogAttrs(context.Background(), slog.LevelError, "docker compose up -d", attrs...)
		return fmt.Errorf("error during docker compose up: %w", err)
	}

	return nil
}

func (c *rootCommand) logInfo(msg string) {
	attrs := []slog.Attr{
		slog.String("url", c.url),
		slog.String("directory", c.directory),
		slog.String("branch", c.branch),
	}

	if c.key != "" {
		attrs = append(attrs, slog.String("key", c.key))
	}

	if c.username != "" {
		attrs = append(attrs, slog.String("username", c.username))
	}

	if c.password != "" {
		attrs = append(attrs, slog.String("password", redacted.Redact(c.password)))
	}

	slog.LogAttrs(context.Background(), slog.LevelInfo, msg, attrs...)
}
