package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// CreateCommand use to implement 'create' command, it create a container.
type CreateCommand struct {
	container
	baseCommand
}

// Init initialize create command.
func (cc *CreateCommand) Init(c *Cli) {
	cc.cli = c

	cc.cmd = cc.init()
	cc.cmd.Use = "create [image]"
	cc.cmd.Short = "Create a new container with specified image"
	cc.cmd.Args = cobra.MinimumNArgs(1)
}

// Run is the entry of create command.
func (cc *CreateCommand) Run(args []string) {
	config := cc.config()
	config.Image = args[0]

	apiClient := cc.cli.Client()
	result, err := apiClient.ContainerCreate(config.ContainerConfig, config.HostConfig, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create container: %v\n", err)
		return
	}

	if len(result.Warnings) != 0 {
		fmt.Printf("WARNING: %s \n", strings.Join(result.Warnings, "\n"))
	}
	fmt.Printf("container's id: %s, name: %s \n", result.ID, result.Name)
}
