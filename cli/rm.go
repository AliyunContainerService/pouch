package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var rmDescription = `
Remove a container object in Pouchd.
If a container be stopped or created, you can remove it. 
If the container be running, you can also remove it with flag force.
When the container be removed, the all resource of the container will
be released.
`

// RmCommand is used to implement 'rm' command.
type RmCommand struct {
	baseCommand
	force bool
}

// Init initializes RmCommand command.
func (r *RmCommand) Init(c *Cli) {
	r.cli = c
	r.cmd = &cobra.Command{
		Use:   "rm [OPTIONS] CONTAINER [CONTAINER...]",
		Short: "Remove one or more containers",
		Long:  rmDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.runRm(args)
		},
		Example: rmExample(),
	}
	r.addFlags()
}

// addFlags adds flags for specific command.
func (r *RmCommand) addFlags() {
	r.cmd.Flags().BoolVarP(&r.force, "force", "f", false, "if the container is running, force to remove it")
}

// runRm is the entry of RmCommand command.
func (r *RmCommand) runRm(args []string) error {
	ctx := context.Background()
	apiClient := r.cli.Client()

	for _, name := range args {
		if err := apiClient.ContainerRemove(ctx, name, r.force); err != nil {
			return fmt.Errorf("failed to remove container: %v", err)
		}
		fmt.Printf("%s\n", name)
	}

	return nil
}

func rmExample() string {
	return `$ pouch rm 5d3152
5d3152

$ pouch rm -f 493028
493028`
}
