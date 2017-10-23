package main

import (
	"github.com/spf13/cobra"
)

// StartCommand use to implement 'start' command, it start a container.
type StartCommand struct {
	baseCommand
}

// Init initialize start command.
func (s *StartCommand) Init(c *Cli) {
	s.cli = c

	s.cmd = &cobra.Command{
		Use:   "start [container]",
		Short: "Start a created container",
		Args:  cobra.MinimumNArgs(1),
	}

	// TODO add flag
}

// Run is the entry of start command.
func (s *StartCommand) Run(args []string) {
	// TODO
}
