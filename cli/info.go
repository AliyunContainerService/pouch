package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alibaba/pouch/apis/types"

	units "github.com/docker/go-units"
	"github.com/spf13/cobra"
)

// infoDespscription is used to describe info command in detail and auto generate command doc.
var infoDescription = "Display the information of pouch, " +
	"including Containers, Images, Storage Driver, Execution Driver, Logging Driver, Kernel Version, " +
	"Operating System, CPUs, Total Memory, Name, ID."

// InfoCommand implements info command.
type InfoCommand struct {
	baseCommand
}

// Init initializes info command.
func (v *InfoCommand) Init(c *Cli) {
	v.cli = c
	v.cmd = &cobra.Command{
		Use:   "info [OPTIONS]",
		Short: "Display system-wide information",
		Args:  cobra.NoArgs,
		Long:  infoDescription,
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

	return prettyPrintInfo(v.cli, result)
}

func prettyPrintInfo(cli *Cli, info *types.SystemInfo) error {
	fmt.Fprintf(os.Stdout, "Containers: %d\n", info.Containers)
	fmt.Fprintf(os.Stdout, " Running: %d\n", info.ContainersRunning)
	fmt.Fprintf(os.Stdout, " Paused: %d\n", info.ContainersPaused)
	fmt.Fprintf(os.Stdout, " Stopped: %d\n", info.ContainersStopped)
	fmt.Fprintf(os.Stdout, "Images: %d\n", info.Images)
	fmt.Fprintf(os.Stdout, "ID: %s\n", info.ID)
	fmt.Fprintf(os.Stdout, "Name: %s\n", info.Name)
	fmt.Fprintf(os.Stdout, "Server Version: %s\n", info.ServerVersion)
	fmt.Fprintf(os.Stdout, "Storage Driver: %s\n", info.Driver)
	fmt.Fprintf(os.Stdout, "Driver Status: %v\n", info.DriverStatus)
	fmt.Fprintf(os.Stdout, "Logging Driver: %s\n", info.LoggingDriver)
	fmt.Fprintf(os.Stdout, "Volume Drivers: %v\n", info.VolumeDrivers)
	fmt.Fprintf(os.Stdout, "Cgroup Driver: %s\n", info.CgroupDriver)
	fmt.Fprintf(os.Stdout, "Default Runtime: %s\n", info.DefaultRuntime)
	if len(info.Runtimes) > 0 {
		fmt.Fprint(os.Stdout, "Runtimes:")
		for name := range info.Runtimes {
			fmt.Fprintf(os.Stdout, " %s", name)
		}
		fmt.Fprint(os.Stdout, "\n")
	}
	fmt.Fprintf(os.Stdout, "runc: %v\n", info.RuncCommit)
	fmt.Fprintf(os.Stdout, "containerd: %v\n", info.ContainerdCommit)

	// Kernel info
	fmt.Fprintf(os.Stdout, "Security Options: %v\n", info.SecurityOptions)
	fmt.Fprintf(os.Stdout, "Kernel Version: %s\n", info.KernelVersion)
	fmt.Fprintf(os.Stdout, "Operating System: %s\n", info.OperatingSystem)
	fmt.Fprintf(os.Stdout, "OSType: %s\n", info.OSType)
	fmt.Fprintf(os.Stdout, "Architecture: %s\n", info.Architecture)

	fmt.Fprintf(os.Stdout, "HTTP Proxy: %s\n", info.HTTPProxy)
	fmt.Fprintf(os.Stdout, "HTTPS Proxy: %s\n", info.HTTPSProxy)
	fmt.Fprintf(os.Stdout, "Registry: %s\n", info.IndexServerAddress)
	fmt.Fprintf(os.Stdout, "Experimental: %v\n", info.ExperimentalBuild)
	fmt.Fprintf(os.Stdout, "Debug: %v\n", info.Debug)
	if len(info.Labels) != 0 {
		fmt.Fprintln(os.Stdout, "Labels:")
		for _, label := range info.Labels {
			fmt.Fprintf(os.Stdout, "  %s\n", label)
		}
	}

	fmt.Fprintf(os.Stdout, "CPUs: %d\n", info.NCPU)
	fmt.Fprintf(os.Stdout, "Total Memory: %s\n", units.BytesSize(float64(info.MemTotal)))
	fmt.Fprintf(os.Stdout, "Pouch Root Dir: %s\n", info.PouchRootDir)
	fmt.Fprintf(os.Stdout, "LiveRestoreEnabled: %v\n", info.LiveRestoreEnabled)
	fmt.Fprintf(os.Stdout, "LxcfsEnabled: %v\n", info.LxcfsEnabled)
	fmt.Fprintf(os.Stdout, "CriEnabled: %v\n", info.CriEnabled)
	if info.RegistryConfig != nil && (len(info.RegistryConfig.InsecureRegistryCIDRs) > 0 || len(info.RegistryConfig.IndexConfigs) > 0) {
		fmt.Fprintln(os.Stdout, "Insecure Registries:")
		for _, registry := range info.RegistryConfig.IndexConfigs {
			if !registry.Secure {
				fmt.Fprintf(os.Stdout, " %s\n", registry.Name)
			}
		}

		for _, registry := range info.RegistryConfig.InsecureRegistryCIDRs {
			fmt.Fprintf(os.Stdout, " %s\n", registry)
		}
	}

	if info.RegistryConfig != nil && len(info.RegistryConfig.Mirrors) > 0 {
		fmt.Fprintln(os.Stdout, "Registry Mirrors:")
		for _, mirror := range info.RegistryConfig.Mirrors {
			fmt.Fprintf(os.Stdout, " %s\n", mirror)
		}
	}

	fmt.Fprintf(os.Stdout, "Daemon Listen Addresses: %v\n", info.ListenAddresses)

	return nil
}

// infoExample shows examples in info command, and is used in auto-generated cli docs.
func infoExample() string {
	return `$ pouch info
Containers: 1
 Running: 1
 Paused: 0
 Stopped: 0
Images:  0
ID:
Name:
Server Version: 0.3-dev
Storage Driver:
Driver Status: []
Logging Driver:
Cgroup Driver:
runc: <nil>
containerd: <nil>
Security Options: []
Kernel Version: 3.10.0-693.17.1.el7.x86_64
Operating System:
OSType: linux
Architecture:
HTTP Proxy: http://127.0.0.1:5678
HTTPS Proxy:
Registry: https://index.docker.io/v1/
Experimental: false
Debug: true
Labels: []
CPUs: 0
Total Memory: 0
Pouch Root Dir: /var/lib/pouch
LiveRestoreEnabled: false
Daemon Listen Addresses: [unix:///var/run/pouchd.sock]
`
}
