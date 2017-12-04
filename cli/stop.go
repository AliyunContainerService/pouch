package main

import (
	"fmt"

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
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.runStop(args)
		},
	}
	s.addFlags()
}

// addFlags adds flags for specific command.
func (s *StopCommand) addFlags() {
	// TODO: add flags here
}

// runStop is the entry of stop command.
func (s *StopCommand) runStop(args []string) error {
	apiClient := s.cli.Client()

	container := args[0]

	if err := apiClient.ContainerStop(container); err != nil {
		return fmt.Errorf("failed to stop container %s: %v", container, err)
	}
	return nil
}
