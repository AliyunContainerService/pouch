package main

import (
	"context"

	"fmt"
	"github.com/spf13/cobra"
)

// restartDescription is used to describe restart command in detail and auto generate command doc.
var restartDescription = "restart one or more containers"

// RestartCommand uses to implement 'restart' command, it restarts one or more containers.
type RestartCommand struct {
	baseCommand
	timeout int
}

// Init initialize restart command.
func (rc *RestartCommand) Init(c *Cli) {
	rc.cli = c

	rc.cmd = &cobra.Command{
		Use:   "restart [OPTION] CONTAINER [CONTAINERS]",
		Short: "restart one or more containers",
		Long:  restartDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return rc.runRestart(args)
		},
		Example: restartExample(),
	}
	rc.addFlags()
}

// addFlags adds flags for specific command.
func (rc *RestartCommand) addFlags() {
	flagSet := rc.cmd.Flags()
	flagSet.IntVarP(&rc.timeout, "time", "t", 10,
		"Seconds to wait for stop before killing the container (default 10)")
}

// runRestart is the entry of restart command.
func (rc *RestartCommand) runRestart(args []string) error {
	ctx := context.Background()
	apiClient := rc.cli.Client()

	for _, name := range args {
		if err := apiClient.ContainerRestart(ctx, name, rc.timeout); err != nil {
			return fmt.Errorf("failed to restart container: %v", err)
		}
		fmt.Printf("%s\n", name)
	}

	return nil
}

// restartExample shows examples in restart command, and is used in auto-generated cli docs.
func restartExample() string {
	return `//TODO`
}
