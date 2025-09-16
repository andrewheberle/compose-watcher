package cmd

import (
	"context"

	"github.com/andrewheberle/simplecommand"
	"github.com/bep/simplecobra"
)

type rootCommand struct {
	// command line args
	url       string
	directory string
	username  string
	password  string
	key       string
	branch    string

	*simplecommand.Command
}

func (c *rootCommand) Init(cd *simplecobra.Commandeer) error {
	if err := c.Command.Init(cd); err != nil {
		return err
	}

	cmd := cd.CobraCommand
	cmd.PersistentFlags().StringVarP(&c.url, "url", "r", "", "URL of Git repository")
	cmd.MarkFlagRequired("url")
	cmd.PersistentFlags().StringVarP(&c.username, "username", "u", "", "Username")
	cmd.PersistentFlags().StringVarP(&c.password, "password", "p", "", "Password for remote repository or SSH private key")
	cmd.MarkFlagsRequiredTogether("username", "password")
	cmd.PersistentFlags().StringVarP(&c.key, "key", "k", "", "SSH private key")
	cmd.MarkFlagsMutuallyExclusive("username", "key")
	cmd.PersistentFlags().StringVarP(&c.branch, "branch", "b", "main", "Git branch to use")
	cmd.PersistentFlags().StringVarP(&c.directory, "directory", "d", "", "Directory for clone/pull")
	cmd.MarkFlagRequired("directory")

	return nil
}

func (c *rootCommand) PreRun(this, runner *simplecobra.Commandeer) error {
	if err := c.Command.PreRun(this, runner); err != nil {
		return err
	}

	c.logInfo("configuration set via cli and environment")

	return nil
}

func (c *rootCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	return cd.CobraCommand.Help()
}

func Execute(ctx context.Context, args []string) error {
	// initial root command
	rootCmd := &rootCommand{
		Command: simplecommand.New(
			"compose-watcher",
			"A CD solution for Docker Compose",
			simplecommand.WithViper("watcher", nil),
		),
	}

	// set up subcommands
	rootCmd.SubCommands = []simplecobra.Commander{
		&initCommand{
			Command: simplecommand.New("init", "Initialise the repository"),
		},
		&checkCommand{
			Command: simplecommand.New("check", "Checks for updates"),
		},
		&watchCommand{
			Command: simplecommand.New("watch", "Watches for changes"),
		},
	}

	x, err := simplecobra.New(rootCmd)
	if err != nil {
		return err
	}

	// run command with the provided args
	if _, err := x.Execute(ctx, args); err != nil {
		return err
	}

	return nil
}
