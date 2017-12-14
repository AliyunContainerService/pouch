package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// createDescription is used to describe create command in detail and auto generate command doc.
var createDescription = "Create a static container object in Pouchd. " +
	"When creating, all configuration user input will be stored in memory store of Pouchd. " +
	"This is useful when you wish to create a container configuration ahead of time so that Pouchd will preserve the resource in advance. " +
	"The container you created is ready to start when you need it."

// CreateCommand use to implement 'create' command, it create a container.
type CreateCommand struct {
	container
	baseCommand
}

// Init initialize create command.
func (cc *CreateCommand) Init(c *Cli) {
	cc.cli = c
	cc.cmd = &cobra.Command{
		Use:   "create [image]",
		Short: "Create a new container with specified image",
		Long:  createDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.runCreate(args)
		},
		Example: createExample(),
	}
	cc.addFlags()
}

// addFlags adds flags for specific command.
func (cc *CreateCommand) addFlags() {
	flagSet := cc.cmd.Flags()
	flagSet.StringVar(&cc.name, "name", "", "Specify name of container")
	flagSet.BoolVarP(&cc.tty, "tty", "t", false, "Allocate a tty device")
	flagSet.StringSliceVarP(&cc.volume, "volume", "v", nil, "Bind mount volumes to container")
	flagSet.StringVar(&cc.runtime, "runtime", "", "Specify oci runtime")
}

// runCreate is the entry of create command.
func (cc *CreateCommand) runCreate(args []string) error {
	config := cc.config()

	config.Image = args[0]
	if len(args) == 2 {
		config.Cmd = strings.Fields(args[1])
	}
	containerName := cc.name

	apiClient := cc.cli.Client()
	result, err := apiClient.ContainerCreate(config.ContainerConfig, config.HostConfig, containerName)
	if err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	if len(result.Warnings) != 0 {
		fmt.Printf("WARNING: %s \n", strings.Join(result.Warnings, "\n"))
	}
	fmt.Printf("container ID: %s, name: %s \n", result.ID, result.Name)
	return nil
}

// createExample shows examples in create command, and is used in auto-generated cli docs.
func createExample() string {
	return `$ pouch create --name foo busybox:latest
container ID: e1d541722d68dc5d133cca9e7bd8fd9338603e1763096c8e853522b60d11f7b9, name: foo`
}
