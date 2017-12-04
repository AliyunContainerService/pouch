package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// InspectCommand is used to implement 'inspect' command.
type InspectCommand struct {
	baseCommand
}

// Init initializes InspectCommand command.
func (p *InspectCommand) Init(c *Cli) {
	p.cli = c
	p.cmd = &cobra.Command{
		Use:   "inspect [container]",
		Short: "get the detailed information of container",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runInpsect(args)
		},
	}
}

// runInpsect is the entry of InspectCommand command.
func (p *InspectCommand) runInpsect(args []string) error {
	apiClient := p.cli.Client()
	name := args[0]
	container, err := apiClient.ContainerGet(name)
	if err != nil {
		return err
	}

	containerjson, err := json.MarshalIndent(&container, "", "  ")
	if err != nil {
		return err
	}
	fmt.Print(string(containerjson) + "\n")
	return nil
}
