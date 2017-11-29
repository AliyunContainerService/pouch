package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// PsCommand is used to implement 'ps' command.
type InspectCommand struct {
	baseCommand
}

// Init initializes InspectCommand command.
func (p *InspectCommand) Init(c *Cli) {
	p.cli = c
	p.cmd = &cobra.Command{
		Use:   "inspect [container]",
		Short: "get the detailed information of containers",
		Args:  cobra.MinimumNArgs(1),
	}
}

// Run is the entry of InspectCommand command.
func (p *InspectCommand) Run(args []string) {
	apiClient := p.cli.Client()
	name := args[0]
	container, err := apiClient.ContainerInfo(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get container: %v\n", err)
		return
	}

	p.cli.Print(container)
}