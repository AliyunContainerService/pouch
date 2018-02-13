package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// unpauseDescription is used to describe unpause command in detail and auto generate command doc.
var unpauseDescription = "Unpause a paused container in Pouchd. " +
	"when unpausing, the paused container will resumes the process execution within the container." +
	"The container you unpaused will be running again if no error occurs."

// UnpauseCommand use to implement 'unpause' command, it unpauses a container.
type UnpauseCommand struct {
	baseCommand
}

// Init initialize unpause command.
func (p *UnpauseCommand) Init(c *Cli) {
	p.cli = c
	p.cmd = &cobra.Command{
		Use:   "unpause CONTAINER",
		Short: "Unpause a paused container",
		Long:  unpauseDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runUnpause(args)
		},
		Example: unpauseExample(),
	}
}

// runUnpause is the entry of unpause command.
func (p *UnpauseCommand) runUnpause(args []string) error {
	ctx := context.Background()
	apiClient := p.cli.Client()

	container := args[0]

	if err := apiClient.ContainerUnpause(ctx, container); err != nil {
		return fmt.Errorf("failed to unpause container %s: %v", container, err)
	}
	return nil
}

// unpauseExample shows examples in unpause command, and is used in auto-generated cli docs.
func unpauseExample() string {
	return `$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Paused   docker.io/library/busybox:latest   runc
$ pouch unpause foo
$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Running    docker.io/library/busybox:latest   runc`
}
