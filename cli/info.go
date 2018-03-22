package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alibaba/pouch/apis/types"

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

	return prettyPrintInfo(v.cli, result)
}

func prettyPrintInfo(cli *Cli, info *types.SystemInfo) error {
	fmt.Fprintln(os.Stdout, "Containers:", info.Containers)
	fmt.Fprintln(os.Stdout, "Running:", info.ContainersRunning)
	fmt.Fprintln(os.Stdout, "Paused:", info.ContainersPaused)
	fmt.Fprintln(os.Stdout, "Stopped:", info.ContainersStopped)
	fmt.Fprintln(os.Stdout, "Images: ", info.Images)
	fmt.Fprintln(os.Stdout, "ID:", info.ID)
	fmt.Fprintln(os.Stdout, "Name:", info.Name)
	fmt.Fprintln(os.Stdout, "Server Version:", info.ServerVersion)
	fmt.Fprintln(os.Stdout, "Storage Driver:", info.Driver)
	fmt.Fprintln(os.Stdout, "Driver Status:", info.DriverStatus)
	fmt.Fprintln(os.Stdout, "Logging Driver:", info.LoggingDriver)
	fmt.Fprintln(os.Stdout, "Cgroup Driver:", info.CgroupDriver)
	if len(info.Runtimes) > 0 {
		fmt.Fprintln(os.Stdout, "Runtimes:")
		for name := range info.Runtimes {
			fmt.Fprintf(os.Stdout, "%s", name)
		}
		fmt.Fprint(os.Stdout, "\n")
		fmt.Fprintln(os.Stdout, "Default Runtime:", info.DefaultRuntime)
	}
	fmt.Fprintln(os.Stdout, "runc:", info.RuncCommit)
	fmt.Fprintln(os.Stdout, "containerd:", info.ContainerdCommit)

	// Kernel info
	fmt.Fprintln(os.Stdout, "Security Options:", info.SecurityOptions)
	fmt.Fprintln(os.Stdout, "Kernel Version:", info.KernelVersion)
	fmt.Fprintln(os.Stdout, "Operating System:", info.OperatingSystem)
	fmt.Fprintln(os.Stdout, "OSType:", info.OSType)
	fmt.Fprintln(os.Stdout, "Architecture:", info.Architecture)

	fmt.Fprintln(os.Stdout, "HTTP Proxy:", info.HTTPProxy)
	fmt.Fprintln(os.Stdout, "HTTPS Proxy:", info.HTTPSProxy)
	fmt.Fprintln(os.Stdout, "Registry:", info.IndexServerAddress)
	fmt.Fprintln(os.Stdout, "Experimental:", info.ExperimentalBuild)
	fmt.Fprintln(os.Stdout, "Debug:", info.Debug)
	fmt.Fprintln(os.Stdout, "Labels:", info.Labels)

	fmt.Fprintln(os.Stdout, "CPUs:", info.NCPU)
	fmt.Fprintln(os.Stdout, "Total Memory:", info.MemTotal)
	fmt.Fprintln(os.Stdout, "Pouch Root Dir:", info.PouchRootDir)
	fmt.Fprintln(os.Stdout, "LiveRestoreEnabled:", info.LiveRestoreEnabled)
	if info.RegistryConfig != nil && (len(info.RegistryConfig.InsecureRegistryCIDRs) > 0 || len(info.RegistryConfig.IndexConfigs) > 0) {
		fmt.Fprintln(os.Stdout, "Insecure Registries:")
		for _, registry := range info.RegistryConfig.IndexConfigs {
			if !registry.Secure {
				fmt.Fprintln(os.Stdout, " "+registry.Name)
			}
		}

		for _, registry := range info.RegistryConfig.InsecureRegistryCIDRs {
			fmt.Fprintf(os.Stdout, " %s\n", registry)
		}
	}

	if info.RegistryConfig != nil && len(info.RegistryConfig.Mirrors) > 0 {
		fmt.Fprintln(os.Stdout, "Registry Mirrors:")
		for _, mirror := range info.RegistryConfig.Mirrors {
			fmt.Fprintln(os.Stdout, " "+mirror)
		}
	}

	fmt.Fprintln(os.Stdout, "Daemon Listen Addresses:", info.ListenAddresses)

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
HTTP Proxy:
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
