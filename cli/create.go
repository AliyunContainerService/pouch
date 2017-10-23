package main

import (
	"github.com/spf13/cobra"
)

// CreateCommand use to implement 'create' command, it create a container.
type CreateCommand struct {
	baseCommand
}

// Init initialize create command.
func (cc *CreateCommand) Init(c *Cli) {
	cc.cli = c

	cc.cmd = &cobra.Command{
		Use:   "create [image]",
		Short: "Create a new container with specify image",
		Args:  cobra.MinimumNArgs(1),
	}

	// TODO add flag
}

// Run is the entry of create command.
func (cc *CreateCommand) Run(args []string) {
	// TODO
}
