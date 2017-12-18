package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// stopDescription is used to describe stop command in detail and auto generate command doc.
var stopDescription = "Stop a running container in Pouchd. " +
	"This is useful when you wish to stop a container.And Pouchd will stop this running container and release the resource. " +
	"The container that you stopped will be closed. "

// StopCommand use to implement 'stop' command, it stops a container.
type StopCommand struct {
	baseCommand
}

// Init initialize stop command.
func (s *StopCommand) Init(c *Cli) {
	s.cli = c
	s.cmd = &cobra.Command{
		Use:   "stop [container]",
		Short: "Stop a running container",
		Long:  stopDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.runStop(args)
		},
		Example: stopExample(),
	}
	s.addFlags()
}

// addFlags adds flags for specific command.
func (s *StopCommand) addFlags() {
	// TODO: add flags here
}

// runStop is the entry of stop command.
func (s *StopCommand) runStop(args []string) error {
	apiClient := s.cli.Client()

	container := args[0]

	if err := apiClient.ContainerStop(container); err != nil {
		return fmt.Errorf("failed to stop container %s: %v", container, err)
	}
	return nil
}

// stopExample shows examples in stop command, and is used in auto-generated cli docs.
func stopExample() string {
	return `$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Running   docker.io/library/busybox:latest   runc
$ pouch stop foo 
$ pouch ps 
Name     ID       Status    Image                              Runtime
foo      71b9c1   Stopped   docker.io/library/busybox:latest   runc`
}
