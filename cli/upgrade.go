package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"

	"github.com/spf13/cobra"
)

// upgradeDescription is used to describe upgrade command in detail and auto generate command doc.
var upgradeDescription = "upgrade is a feature to replace a container's image. " +
	"You can specify the new Entrypoint and Cmd for the new container. When you want to update " +
	"a container's image, but inherit the network and volumes of the old container, then you should " +
	"think about the upgrade feature."

// UpgradeCommand use to implement 'upgrade' command, it is used to upgrade a container.
type UpgradeCommand struct {
	baseCommand
	entrypoint string
	image      string
}

// Init initialize upgrade command.
func (ug *UpgradeCommand) Init(c *Cli) {
	ug.cli = c
	ug.cmd = &cobra.Command{
		Use:   "upgrade [OPTIONS] CONTAINER [COMMAND] [ARG...]",
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
	flagSet.StringVar(&ug.entrypoint, "entrypoint", "", "Overwrite the default ENTRYPOINT of the image")
	flagSet.StringVar(&ug.image, "image", "", "Specify image of the new container")
}

// runUpgrade is the entry of UpgradeCommand command.
func (ug *UpgradeCommand) runUpgrade(args []string) error {
	var cmd []string

	name := args[0]
	if name == "" {
		return fmt.Errorf("failed to upgrade container: must specify container name")
	}
	if len(args) > 1 {
		cmd = args[1:]
	}

	image := ug.image
	if image == "" {
		return fmt.Errorf("failed to upgrade container: must specify new image")
	}

	upgradeConfig := &types.ContainerUpgradeConfig{
		Image:      image,
		Cmd:        cmd,
		Entrypoint: strings.Fields(ug.entrypoint),
	}

	ctx := context.Background()
	apiClient := ug.cli.Client()

	if err := pullMissingImage(ctx, apiClient, image, false); err != nil {
		return err
	}

	if err := apiClient.ContainerUpgrade(ctx, name, upgradeConfig); err != nil {
		return err
	}

	fmt.Println(name)
	return nil
}

//upgradeExample shows examples in exec command, and is used in auto-generated cli docs.
func upgradeExample() string {
	return ` $ pouch run -d -m 20m --name test  registry.hub.docker.com/library/busybox:latest
4c58d27f58d38776dda31c01c897bbf554c802a9b80ae4dc20be1337f8a969f2
$ pouch upgrade --image registry.hub.docker.com/library/hello-world:latest test
test`
}
