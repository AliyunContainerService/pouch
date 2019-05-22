package main

import (
	"context"
	"fmt"

	"github.com/alibaba/pouch/apis/filters"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/spf13/cobra"
)

// imagePruneDescription is used to delete all unused images.
var imagePruneDescription = "Delete all unused images"

// ImagePruneCommand use to delete all stopped images.
type ImagePruneCommand struct {
	baseCommand

	force      bool
	flagFilter []string
}

// Init initialize "image prune" command.
func (i *ImagePruneCommand) Init(c *Cli) {
	i.cli = c
	i.cmd = &cobra.Command{
		Use:   "prune [OPTIONS]",
		Short: "Delete all unused images",
		Long:  imagePruneDescription,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return i.runImagePrune(args)
		},
		Example: i.example(),
	}
	i.addFlags()
}

// addFlags adds flags for specific command.
func (i *ImagePruneCommand) addFlags() {
	flagSet := i.cmd.Flags()

	flagSet.BoolVarP(&i.force, "force", "f", false, "Do not prompt for confirmation")
	flagSet.StringSliceVarP(&i.flagFilter, "filter", "", []string{}, "Filter output based on conditions provided, filter support reference, since, before")
}

// runImagePrune is used to delete unused images.
func (i *ImagePruneCommand) runImagePrune(args []string) error {
	ctx := context.Background()
	apiClient := i.cli.Client()

	imageFilterArgs, err := filters.FromFilterOpts(i.flagFilter)

	if err != nil {
		return err
	}

	if !i.force {
		fmt.Println("WARNING! This will delete all dangling images..")
		fmt.Print("Are you sure you want to continue? [y/N]")
		var input string
		fmt.Scanf("%s", &input)
		if input != "Y" && input != "y" {
			fmt.Println("Total reclaimed space: 0B")
			return nil
		}
	}

	resp, err := apiClient.ImagePrune(ctx, imageFilterArgs)
	if err != nil {
		return err
	}

	for _, imgPruneInfo := range resp.ImagesDeleted {
		fmt.Println("untagged:")
		fmt.Print(imgPruneInfo.Untagged)
		fmt.Println("Deleted Images:")
		fmt.Print(imgPruneInfo.Deleted)
	}

	fmt.Printf("\nTotal reclaimed space: %s\n", utils.FormatSize(resp.SpaceReclaimed))
	return nil
}

// this shows image prune in images command, and is used in auto-generated cli docs.
func (i *ImagePruneCommand) example() string {
	return `$ pouch image prune
WARNING! This will delete all dangling images.
untagged:
registry.hub.docker.com/library/hello-world:latest
Deleted Images:
sha256:fce289e99eb9bca977dae136fbe2a82b6b7d4c372474c9235adc1741675f587e
sha256:af0b15c8625bb1938f1d7b17081031f649fd14e6b233688eea3c5483994a66a3
untagged:
registry.hub.docker.com/library/busybox:latest
Deleted Images:
sha256:64f5d945efcc0f39ab11b3cd4ba403cc9fefe1fa3613123ca016cf3708e8cafb
sha256:d1156b98822dccbb924b4e5fe16465a7ecac8bfc81d726177bed403a8e70c972

Total reclaimed space: 764841 B
`
}
