package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/alibaba/pouch/pkg/reference"

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

	ctx := context.Background()
	apiClient := ug.cli.Client()

	// Check whether the image has been pulled
	_, err = apiClient.ImageInspect(ctx, config.Image)
	if err != nil && strings.Contains(err.Error(), "not found") {
		fmt.Printf("Image %s not found, try to pull it...\n", config.Image)

		namedRef, err := reference.ParseNamedReference(args[0])
		if err != nil {
			return fmt.Errorf("failed to pull image: %v", err)
		}
		taggedRef := reference.WithDefaultTagIfMissing(namedRef).(reference.Tagged)

		responseBody, err := apiClient.ImagePull(ctx, taggedRef.Name(), taggedRef.Tag(), fetchRegistryAuth(taggedRef.Name()))
		if err != nil {
			return fmt.Errorf("failed to pull image: %v", err)
		}
		defer responseBody.Close()

		if err := showProgress(responseBody); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	containerName := ug.name
	if containerName == "" {
		return fmt.Errorf("failed to upgrade container: must specify container name")
	}

	// TODO if error is image not found, we can pull image, and retry upgrade
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
