package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NetworkCommand is used to implent 'network' command.
type NetworkCommand struct {
	baseCommand
}

// Init initializes NetworkCommand command.
func (n *NetworkCommand) Init(c *Cli) {
	n.cli = c

	n.cmd = &cobra.Command{
		Use:	"network [command]",
		Short:	"Manage pouch networks",
		Args:	cobra.MinimumNArgs(1),
	}

	c.AddCommand(n, &NetworkListCommand{})
}

// RunE is the entry of NetworkCommand command.
func (n *NetworkCommand) RunE(args []string) error {
	return nil
}

// NetworkListCommand is used to implement 'network ls' command.
type NetworkListCommand struct {
	baseCommand
}

// Init initializes NetworkListCommand command.
func (n *NetworkListCommand) Init(c *Cli) {
	n.cli = c

	n.cmd = &cobra.Command{
		Use:	"ls [OPTIONS]",
		Short:	"List networks",
		Args:	cobra.NoArgs,
		RunE:	func(cmd *cobra.Command, args []string) error {
			return n.runNetworkList(args)
		},
	}
	n.addFlags()
}

// addFlags adds flags for specific command.
func (n *NetworkListCommand) addFlags() {
	// TODO: add flags here
}

// runNetworkList is the entry of NetworkListCommand command.
func (n *NetworkListCommand) runNetworkList(args []string) error {
	apiClient := n.cli.Client()

	networks, err := apiClient.NetworkList()
	if err != nil {
		return fmt.Errorf("failed to get network list: %v", err)
	}

	display := n.cli.NewTableDisplay()
	display.AddRow([]string{"Network Id", "Name", "Driver", "Scope"})
	for _, n := range networks {
		display.AddRow([]string{})
	}
}
