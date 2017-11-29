package main

import (
	"os"
)

func main() {
	cli := NewCli().SetFlags()

	base := &baseCommand{cmd: cli.rootCmd, cli: cli}

	// Add all subcommands.
	cli.AddCommand(base, &PullCommand{})
	cli.AddCommand(base, &CreateCommand{})
	cli.AddCommand(base, &StartCommand{})
	cli.AddCommand(base, &StopCommand{})
	cli.AddCommand(base, &PsCommand{})
	cli.AddCommand(base, &ExecCommand{})
	cli.AddCommand(base, &VersionCommand{})
	cli.AddCommand(base, &ImageCommand{})
	cli.AddCommand(base, &VolumeCommand{})
	cli.AddCommand(base, &InspectCommand{})

	if err := cli.Run(); err != nil {
		os.Exit(1)
	}
}
