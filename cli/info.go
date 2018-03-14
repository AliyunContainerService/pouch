package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// infoDespscription is used to describe info command in detail and auto generate command doc.
var infoDescription = "Display the information of pouch, " +
	"including Containers, Images, Storage Driver, Execution Driver, Logging Driver, Kernel Version," +
	"Operating System, CPUs, Total Memory, Name, ID."

// InfoCommand use to implement 'info' command.
type InfoCommand struct {
	baseCommand
}

// Init initialize info command.
func (v *InfoCommand) Init(c *Cli) {
	v.cli = c
	v.cmd = &cobra.Command{
		Use:   "info [OPTIONS]",
		Short: "Display system-wide information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return v.runInfo()
		},
		Example: infoExample(),
	}
	v.addFlags()
}

// addFlags adds flags for specific command.
func (v *InfoCommand) addFlags() {
	// TODO: add flags here
}

// runInfo is the entry of info command.
func (v *InfoCommand) runInfo() error {
	ctx := context.Background()
	apiClient := v.cli.Client()

	result, err := apiClient.SystemInfo(ctx)
	if err != nil {
		return fmt.Errorf("failed to get system info: %v", err)
	}

	v.cli.Print(result)
	return nil
}

// infoExample shows examples in info command, and is used in auto-generated cli docs.
func infoExample() string {
	return `$ pouch info
ID:
Name:
OperatingSystem:
PouchRootDir:         /var/lib/pouch
ServerVersion:        0.3-dev
ContainersRunning:    0
Debug:                false
DriverStatus:         []
Labels:               []
Containers:           0
DefaultRuntime:       runc
Driver:
ExperimentalBuild:    false
KernelVersion:        3.10.0-693.11.6.el7.x86_64
OSType:               linux
CgroupDriver:
ContainerdCommit:     <nil>
ContainersPaused:     0
LoggingDriver:
SecurityOptions:      []
NCPU:                 0
RegistryConfig:       <nil>
RuncCommit:           <nil>
ContainersStopped:    0
HTTPSProxy:
IndexServerAddress:   https://index.docker.io/v1/
LiveRestoreEnabled:   false
Runtimes:             map[]
Architecture:
HTTPProxy:
Images:               0
MemTotal:             0
`
}
