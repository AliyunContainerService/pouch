package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// pauseDescription is used to describe pause command in detail and auto generate command doc.
var pauseDescription = "Pause a running container object in Pouchd. " +
	"when pausing, the container will pause its running but hold all the relevant resource." +
	"This is useful when you wish to pause a container for a while and to restore the running status later." +
	"The container you paused will pause without being terminated."

// PauseCommand use to implement 'pause' command, it pauses a container.
type PauseCommand struct {
	baseCommand
}

// Init initialize pause command.
func (p *PauseCommand) Init(c *Cli) {
	p.cli = c
	p.cmd = &cobra.Command{
		Use:   "pause CONTAINER",
		Short: "Pause a running container",
		Long:  pauseDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runPause(args)
		},
		Example: pauseExample(),
	}
	p.addFlags()
}

// addFlags adds flags for specific command.
func (p *PauseCommand) addFlags() {
	// TODO: add flags here
}

// runPause is the entry of pause command.
func (p *PauseCommand) runPause(args []string) error {
	apiClient := p.cli.Client()

	container := args[0]

	if err := apiClient.ContainerPause(container); err != nil {
		return fmt.Errorf("failed to pause container %s: %v", container, err)
	}
	return nil
}

// pauseExample shows examples in pause command, and is used in auto-generated cli docs.
func pauseExample() string {
	return `$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Running   docker.io/library/busybox:latest   runc
$ pouch pause foo
$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Paused    docker.io/library/busybox:latest   runc`
}
