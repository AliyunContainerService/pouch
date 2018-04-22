package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// stopDescription is used to describe stop command in detail and auto generate command doc.
var stopDescription = "Stop one or more running containers in Pouchd. Waiting the given number of seconds before forcefully killing the container." +
	"This is useful when you wish to stop a container. And Pouchd will stop this running container and release the resource. " +
	"The container that you stopped will be terminated. "

// StopCommand use to implement 'stop' command, it stops a container.
type StopCommand struct {
	baseCommand
	timeout int
}

// Init initialize stop command.
func (s *StopCommand) Init(c *Cli) {
	s.cli = c
	s.cmd = &cobra.Command{
		Use:   "stop [OPTIONS] CONTAINER [CONTAINER...]",
		Short: "Stop one or more running containers",
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
	flagSet := s.cmd.Flags()
	flagSet.IntVarP(&s.timeout, "time", "t", 10, "Seconds to wait for stop before killing it")
}

// runStop is the entry of stop command.
func (s *StopCommand) runStop(args []string) error {
	ctx := context.Background()
	apiClient := s.cli.Client()

	var errs []string
	for _, name := range args {
		if err := apiClient.ContainerStop(ctx, name, strconv.Itoa(s.timeout)); err != nil {
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
