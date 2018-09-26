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
	base.Cmd().SilenceErrors = true

	// Add all subcommands.
	cli.AddCommand(base, &PullCommand{})
	cli.AddCommand(base, &CreateCommand{})
	cli.AddCommand(base, &StartCommand{})
	cli.AddCommand(base, &StopCommand{})
	cli.AddCommand(base, &PsCommand{})
	cli.AddCommand(base, &RmCommand{})
	cli.AddCommand(base, &RestartCommand{})
	cli.AddCommand(base, &ExecCommand{})
	cli.AddCommand(base, &VersionCommand{})
	cli.AddCommand(base, &InfoCommand{})
	cli.AddCommand(base, &ImageMgmtCommand{})
	cli.AddCommand(base, &ImagesCommand{})
	cli.AddCommand(base, &RmiCommand{})
	cli.AddCommand(base, &VolumeCommand{})
	cli.AddCommand(base, &NetworkCommand{})
	cli.AddCommand(base, &TagCommand{})
	cli.AddCommand(base, &LoadCommand{})
	cli.AddCommand(base, &SaveCommand{})
	cli.AddCommand(base, &HistoryCommand{})

	cli.AddCommand(base, &InspectCommand{})
	cli.AddCommand(base, &RenameCommand{})
	cli.AddCommand(base, &PauseCommand{})
	cli.AddCommand(base, &UnpauseCommand{})
	cli.AddCommand(base, &RunCommand{})
	cli.AddCommand(base, &LoginCommand{})
	cli.AddCommand(base, &UpdateCommand{})
	cli.AddCommand(base, &LogoutCommand{})
	cli.AddCommand(base, &UpgradeCommand{})
	cli.AddCommand(base, &TopCommand{})
	cli.AddCommand(base, &LogsCommand{})
	cli.AddCommand(base, &RemountLxcfsCommand{})
	cli.AddCommand(base, &WaitCommand{})
	cli.AddCommand(base, &DaemonUpdateCommand{})
	cli.AddCommand(base, &CheckpointCommand{})
	cli.AddCommand(base, &EventsCommand{})
	cli.AddCommand(base, &CommitCommand{})
	cli.AddCommand(base, &CopyCommand{})

	// add generate doc command
	cli.AddCommand(base, &GenDocCommand{})

	if err := cli.Run(); err != nil {
		// deal with ExitError, which should be recognize as error, and should
		// not be exit with status 0.
		if exitErr, ok := err.(ExitError); ok {
			if exitErr.Status != "" {
				fmt.Fprintln(os.Stderr, exitErr.Status)
			}
			if exitErr.Code == 0 {
				// when get error with ExitError, code should not be 0.
				exitErr.Code = 1
			}
			os.Exit(exitErr.Code)
		}

		// not ExitError, print error to os.Stderr, exit code 1.
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
