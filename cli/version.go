package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// versionDescription is used to describe version command in detail and auto generate command doc.
var versionDescription = "Display the version information of pouch client and daemon， " +
	"including GoVersion, KernelVersion, Os, Version, APIVersion, Arch, BuildTime and GitCommit."

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
		Example: versionExample(),
	}
	v.addFlags()
}

// addFlags adds flags for specific command.
func (v *VersionCommand) addFlags() {
	// TODO: add flags here
}

// runVersion is the entry of version command.
func (v *VersionCommand) runVersion() error {
	ctx := context.Background()
	apiClient := v.cli.Client()

	result, err := apiClient.SystemVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get system version: %v", err)
	}

	v.cli.Print(result)
	return nil
}

// versionExample shows examples in version command, and is used in auto-generated cli docs.
func versionExample() string {
	return `$ pouch version
GoVersion:       go1.9.1
KernelVersion:   3.10.0-693.11.6.el7.x86_64
Os:              linux
Version:         0.1.0-dev
APIVersion:      1.24
Arch:            amd64
BuildTime:       2017-12-18T07:48:56.348129663Z
GitCommit:
`
}
