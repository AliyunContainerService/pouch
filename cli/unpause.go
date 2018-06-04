package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// unpauseDescription is used to describe unpause command in detail and auto generate command doc.
var unpauseDescription = "Unpause one or more paused containers in Pouchd. " +
	"when unpausing, the paused container will resumes the process execution within the container." +
	"The container you unpaused will be running again if no error occurs."

// UnpauseCommand use to implement 'unpause' command, it unpauses one or more containers.
type UnpauseCommand struct {
	baseCommand
}

// Init initialize unpause command.
func (p *UnpauseCommand) Init(c *Cli) {
	p.cli = c
	p.cmd = &cobra.Command{
		Use:   "unpause CONTAINER [CONTAINER...]",
		Short: "Unpause one or more paused container",
		Long:  unpauseDescription,
		Args:  cobra.MinimumNArgs(1),
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

	var errs []string
	for _, name := range args {
		if err := apiClient.ContainerUnpause(ctx, name); err != nil {
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

// unpauseExample shows examples in unpause command, and is used in auto-generated cli docs.
func unpauseExample() string {
	return `$ pouch ps
Name   ID       Status                  Created          Image                                            Runtime
foo2   c95673   Up 13 seconds(paused)   14 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
foo1   204cc6   Up 17 seconds(paused)   17 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch unpause foo1 foo2
foo1
foo2
$ pouch ps
Name   ID       Status          Created          Image                                            Runtime
foo2   c95673   Up 48 seconds   49 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
foo1   204cc6   Up 52 seconds   52 seconds ago   registry.hub.docker.com/library/busybox:latest   runc`
}
