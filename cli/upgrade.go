package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// upgradeDescription is used to describe upgrade command in detail and auto generate command doc.
var upgradeDescription = ""

// UpgradeCommand use to implement 'upgrade' command, it is used to upgrade a container.
type UpgradeCommand struct {
	baseCommand
	*container
}

// Init initialize upgrade command.
func (ug *UpgradeCommand) Init(c *Cli) {
	ug.cli = c
	ug.cmd = &cobra.Command{
		Use:   "upgrade [OPTIONS] IMAGE [COMMAND] [ARG...]",
		Short: "Upgrade a container with new image and args",
		Long:  upgradeDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ug.runUpgrade(args)
		},
		Example: upgradeExample(),
	}
	ug.addFlags()
}

// addFlags adds flags for specific command.
func (ug *UpgradeCommand) addFlags() {
	flagSet := ug.cmd.Flags()
	flagSet.SetInterspersed(false)

	c := addCommonFlags(flagSet)
	ug.container = c
}

// runUpgrade is the entry of UpgradeCommand command.
func (ug *UpgradeCommand) runUpgrade(args []string) error {
	config, err := ug.config()
	if err != nil {
		return fmt.Errorf("failed to upgrade container: %v", err)
	}

	config.Image = args[0]
	if len(args) > 1 {
		config.Cmd = args[1:]
	}

	containerName := ug.name
	if containerName == "" {
		return fmt.Errorf("failed to upgrade container: must specify container name")
	}

	ctx := context.Background()
	apiClient := ug.cli.Client()

	if err := pullMissingImage(ctx, apiClient, config.Image, false); err != nil {
		return err
	}

	err = apiClient.ContainerUpgrade(ctx, containerName, config.ContainerConfig, config.HostConfig)
	if err == nil {
		fmt.Println(containerName)
	}

	return err
}

//upgradeExample shows examples in exec command, and is used in auto-generated cli docs.
func upgradeExample() string {
	return ` $ pouch run -d -m 20m --name test1  registry.hub.docker.com/library/busybox:latest
4c58d27f58d38776dda31c01c897bbf554c802a9b80ae4dc20be1337f8a969f2
$ pouch upgrade --name test1 registry.hub.docker.com/library/hello-world:latest
test1`
}
