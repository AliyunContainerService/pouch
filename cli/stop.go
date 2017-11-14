package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// StopCommand use to implement 'stop' command, it stops a container.
type StopCommand struct {
	baseCommand
}

// Init initialize stop command.
func (s *StopCommand) Init(c *Cli) {
	s.cli = c

	s.cmd = &cobra.Command{
		Use:   "stop [container]",
		Short: "Stop a running container",
		Args:  cobra.MinimumNArgs(1),
	}

	// TODO add flag
}

// Run is the entry of stop command.
func (s *StopCommand) Run(args []string) {
	apiClient := s.cli.Client()

	container := args[0]

	if err := apiClient.ContainerStop(container); err != nil {
		fmt.Fprintf(os.Stderr, "failed to stop container %s: %v \n", container, err)
		return
	}
}
