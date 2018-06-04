package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"

	"github.com/spf13/cobra"
)

var rmDescription = `
Remove one or more containers in Pouchd.
If a container be stopped or created, you can remove it. 
If the container is running, you can also remove it with flag force.
When the container is removed, the all resources of the container will
be released.
`

// RmCommand is used to implement 'rm' command.
type RmCommand struct {
	baseCommand
	force         bool
	removeVolumes bool
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
	flagSet := r.cmd.Flags()

	flagSet.BoolVarP(&r.force, "force", "f", false, "if the container is running, force to remove it")
	flagSet.BoolVarP(&r.removeVolumes, "volumes", "v", false, "remove container's volumes that create by the container")
}

// runRm is the entry of RmCommand command.
func (r *RmCommand) runRm(args []string) error {
	ctx := context.Background()
	apiClient := r.cli.Client()

	options := &types.ContainerRemoveOptions{
		Force:   r.force,
		Volumes: r.removeVolumes,
	}

	var errs []string
	for _, name := range args {
		if err := apiClient.ContainerRemove(ctx, name, options); err != nil {
			errs = append(errs, err.Error())
			continue
		}
		fmt.Printf("%s\n", name)
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

func rmExample() string {
	return `$ pouch ps -a
Name   ID       Status                  Created          Image                                            Runtime
foo    03cd58   Exited (0) 25 seconds   26 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch rm foo
foo
$ pouch ps
Name   ID       Status         Created          Image                                            Runtime
foo2   1d979d   Up 5 seconds   6 seconds ago    registry.hub.docker.com/library/busybox:latest   runc
foo1   83e3cf   Up 9 seconds   10 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch rm -f foo1 foo2
foo1
foo2`
}
