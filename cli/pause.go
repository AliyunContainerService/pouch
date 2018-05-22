package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// pauseDescription is used to describe pause command in detail and auto generate command doc.
var pauseDescription = "Pause one or more running containers in Pouchd. " +
	"when pausing, the container will pause its running but hold all the relevant resource." +
	"This is useful when you wish to pause a container for a while and to restore the running status later." +
	"The container you paused will pause without being terminated."

// PauseCommand use to implement 'pause' command, it pauses one or more containers.
type PauseCommand struct {
	baseCommand
}

// Init initialize pause command.
func (p *PauseCommand) Init(c *Cli) {
	p.cli = c
	p.cmd = &cobra.Command{
		Use:   "pause CONTAINER [CONTAINER...]",
		Short: "Pause one or more running containers",
		Long:  pauseDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runPause(args)
		},
		Example: pauseExample(),
	}
}

// runPause is the entry of pause command.
func (p *PauseCommand) runPause(args []string) error {
	ctx := context.Background()
	apiClient := p.cli.Client()

	var errs []string
	for _, name := range args {
		if err := apiClient.ContainerPause(ctx, name); err != nil {
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

// pauseExample shows examples in pause command, and is used in auto-generated cli docs.
func pauseExample() string {
	return `$ pouch ps
Name   ID       Status          Created          Image                                            Runtime
foo2   87259c   Up 25 seconds   26 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
foo1   77188c   Up 46 seconds   47 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch pause foo1 foo2
foo1
foo2
$ pouch ps
Name   ID       Status                Created        Image                                            Runtime
foo2   87259c   Up 1 minute(paused)   1 minute ago   registry.hub.docker.com/library/busybox:latest   runc
foo1   77188c   Up 1 minute(paused)   1 minute ago   registry.hub.docker.com/library/busybox:latest   runc`
}
