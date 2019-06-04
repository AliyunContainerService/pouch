package main

import (
	"context"
	"fmt"

	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/pkg/utils/filters"

	"github.com/spf13/cobra"
)

// containerPruneDescription is used to delete all stopped containers.
var containerPruneDescription = "Delete all stopped containers"

// ContainerPruneCommand use to delete all stopped containers.
type ContainerPruneCommand struct {
	baseCommand

	force      bool
	flagFilter []string
}

// Init initialize "container prune" command.
func (i *ContainerPruneCommand) Init(c *Cli) {
	i.cli = c
	i.cmd = &cobra.Command{
		Use:   "prune [OPTIONS]",
		Short: "Delete all stopped containers",
		Long:  containerPruneDescription,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return i.runContainerPrune(args)
		},
		Example: i.example(),
	}
	i.addFlags()
}

// addFlags adds flags for specific command.
func (i *ContainerPruneCommand) addFlags() {
	flagSet := i.cmd.Flags()

	flagSet.BoolVarP(&i.force, "force", "f", false, "Do not prompt for confirmation")
	flagSet.StringSliceVarP(&i.flagFilter, "filter", "", nil, "Provide filter values (e.g. 'until=<timestamp>')")
}

// runContainerPrune is used to delete unused containers.
func (i *ContainerPruneCommand) runContainerPrune(args []string) error {
	ctx := context.Background()
	apiClient := i.cli.Client()

	filter, err := filters.Parse(i.flagFilter)
	if err != nil {
		return err
	}

	if !i.force {
		fmt.Println("WARNING! This will delete all stopped containers.")
		fmt.Print("Are you sure you want to continue? [y/N]")
		var input string
		fmt.Scanf("%s", &input)
		if input != "Y" && input != "y" {
			fmt.Println("Total reclaimed space: 0B")
			return nil
		}
	}

	resp, err := apiClient.ContainerPrune(ctx, filter)

	if err != nil {
		return err
	}

	// print resp info
	if len(resp.ContainersDeleted) > 0 {
		fmt.Println("Deleted Containers:")
	}
	for _, containerDeleted := range resp.ContainersDeleted {
		fmt.Println(containerDeleted)
	}

	fmt.Printf("\nTotal reclaimed space: %s\n", utils.FormatSize(resp.SpaceReclaimed))
	return nil
}

// this shows examples in container prune command, and is used in auto-generated cli docs.
func (i *ContainerPruneCommand) example() string {
	return `$ pouch container prune
WARNING! This will delete all stopped containers.
Are you sure you want to continue? [y/N]y
Deleted Containers:
9beaac36a373b7f0bf9d26887cc4cc16b4b287ec4d411da98b7220ea985f5ed8

Total reclaimed space: 34B`
}
