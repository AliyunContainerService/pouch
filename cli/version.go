package main

import (
	"fmt"
	"os"

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
		Short: "Print version",
	}
}

// Run is the entry of version command.
func (v *VersionCommand) Run(args []string) {
	apiClient := v.cli.Client()

	result, err := apiClient.SystemVersion()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get system version: %v\n", err)
		return
	}

	v.cli.Print(result)
}
