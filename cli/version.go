package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// VersionCommand use to implement 'version' command.
type VersionCommand struct {
	baseCommand
}

// Init initialize version command.
func (v *VersionCommand) Init(c *Cli) {
	v.cli = c
	v.cmd = &cobra.Command{
		Use:   "version",
		Short: "Print versions about Pouch CLI and Pouchd",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.runVersion()
		},
	}
	v.addFlags()
}

// addFlags adds flags for specific command.
func (v *VersionCommand) addFlags() {
	// TODO: add flags here
}

// runVersion is the entry of version command.
func (v *VersionCommand) runVersion() error {
	apiClient := v.cli.Client()

	result, err := apiClient.SystemVersion()
	if err != nil {
		return fmt.Errorf("failed to get system version: %v", err)
	}

	v.cli.Print(result)
	return nil
}
