package main

import (
	"fmt"
	"os"
)

func main() {
	cli := NewCli()

	// set global flags for rootCmd in cli.
	cli.SetFlags()

	base := &baseCommand{cmd: cli.rootCmd, cli: cli}

	// Add all subcommands.
	cli.AddCommand(base, &PullCommand{})
	cli.AddCommand(base, &CreateCommand{})
	cli.AddCommand(base, &StartCommand{})
	cli.AddCommand(base, &StopCommand{})
	cli.AddCommand(base, &PsCommand{})
	cli.AddCommand(base, &RmCommand{})
	cli.AddCommand(base, &ExecCommand{})
	cli.AddCommand(base, &VersionCommand{})
	cli.AddCommand(base, &ImageCommand{})
	cli.AddCommand(base, &RmiCommand{})
	cli.AddCommand(base, &VolumeCommand{})
	cli.AddCommand(base, &InspectCommand{})
	cli.AddCommand(base, &RenameCommand{})
	cli.AddCommand(base, &PauseCommand{})

	// add generate doc command
	cli.AddCommand(base, &GenDocCommand{})

	if err := cli.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
