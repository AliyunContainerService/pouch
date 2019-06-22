package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// killDescription is used to describe kill command in detail and auto generate command doc.
var killDescription = "Kill one or more running container objects in Pouchd. " +
	"You can kill a container using the container’s ID, ID-prefix, or name. " +
	"This is useful when you wish to kill a container which is running."

// KillCommand use to implement 'kill' command, it kill one or more containers.
type KillCommand struct {
	baseCommand
	signal string
}

// Init initialize kill command.
func (s *KillCommand) Init(c *Cli) {
	s.cli = c
	s.cmd = &cobra.Command{
		Use:   "kill [OPTIONS] CONTAINER [CONTAINER...]",
		Short: "kill one or more running containers",
		Long:  killDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.runKill(args)
		},
		Example: killExample(),
	}
	s.addFlags()
}

// addFlags adds flags for specific command.
func (s *KillCommand) addFlags() {
	flagSet := s.cmd.Flags()
	flagSet.StringVarP(&s.signal, "signal", "s", "KILL", "Signal to send to the container")

}

// runKill is the entry of kill command.
func (s *KillCommand) runKill(args []string) error {
	ctx := context.Background()
	apiClient := s.cli.Client()

	var errs []string
	for _, name := range args {
		if err := apiClient.ContainerKill(ctx, name, s.signal); err != nil {
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

// killExample shows examples in kill command, and is used in auto-generated cli docs.
func killExample() string {
	return `$ pouch ps -a
Name   ID       Status    		Created         Image                                            Runtime
foo2   5a0ede   Up 2 seconds   	3 second ago    registry.hub.docker.com/library/busybox:latest   runc
foo1   e05637   Up 6 seconds   	7 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch kill foo1
foo1
$ pouch ps
Name   ID       Status         Created          Image                                            Runtime
foo2   5a0ede   Up 11 seconds   12 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
`
}
