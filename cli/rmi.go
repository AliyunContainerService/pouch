package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rmiDescription = "Remove one or more images by reference"

// RmiCommand use to implement 'rmi' command, it remove one or more images by reference
type RmiCommand struct {
	baseCommand
	force bool
}

// Init initialize rmi command
func (rmi *RmiCommand) Init(c *Cli) {
	rmi.cli = c
	rmi.cmd = &cobra.Command{
		Use:   "rmi image ",
		Short: "Remove one or more images by reference",
		Long:  rmiDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return rmi.runRmi(args)
		},
		Example: rmiExample(),
	}
	rmi.addFlags()
}

// addFlags adds flags for specific command
func (rmi *RmiCommand) addFlags() {
	rmi.cmd.Flags().BoolVarP(&rmi.force, "force", "f", false, "if image is being used, remove image and all associated resources")
}

// runRmi is the entry of rmi command
func (rmi *RmiCommand) runRmi(args []string) error {
	apiClient := rmi.cli.Client()

	for _, name := range args {
		if err := apiClient.ImageRemove(name, rmi.force); err != nil {
			return fmt.Errorf("failed to remove image: %v", err)
		}
		fmt.Printf("%s\n", name)
	}

	return nil
}

// rmiExample shows examples in rmi command, and is used in auto-generated cli docs.
func rmiExample() string {
	return `$ pouch rmi docker.io/library/busybox:latest
docker.io/library/busybox:latest`
}
