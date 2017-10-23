package main

import (
	"github.com/spf13/cobra"
)

// PullCommand use to implement 'pull' command, it download image.
type PullCommand struct {
	baseCommand
}

// Init initialize pull command.
func (p *PullCommand) Init(c *Cli) {
	p.cli = c

	p.cmd = &cobra.Command{
		Use:   "pull [image]",
		Short: "Pull use to download image from repository",
		Args:  cobra.MinimumNArgs(1),
	}
}

// Run is the entry of pull command.
func (p *PullCommand) Run(args []string) {
	// TODO
	name := args[0]

	println(name)
}
