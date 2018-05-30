package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// waitDescription is used to describe wait command in detail and auto generate command doc.
var waitDescription = "Block until one or more containers stop, then print their exit codes. " +
	"If container state is already stopped, the command will return exit code immediately. " +
	"On a successful stop, the exit code of the container is returned. "

// WaitCommand is used to implement 'wait' command.
type WaitCommand struct {
	baseCommand
}

// Init initializes wait command.
func (wait *WaitCommand) Init(c *Cli) {
	wait.cli = c
	wait.cmd = &cobra.Command{
		Use:   "wait CONTAINER [CONTAINER...]",
		Short: "Block until one or more containers stop, then print their exit codes",
		Long:  waitDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return wait.runWait(args)
		},
		Example: waitExamples(),
	}
}

// runWait is the entry of wait command.
func (wait *WaitCommand) runWait(args []string) error {
	ctx := context.Background()
	apiClient := wait.cli.Client()

	var errs []string
	for _, name := range args {
		response, err := apiClient.ContainerWait(ctx, name)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		fmt.Printf("%d\n", response.StatusCode)
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}

	return nil
}

// waitExamples shows examples in wait command, and is used in auto-generated cli docs.
func waitExamples() string {
	return `$ pouch ps
Name   ID       Status         Created         Image                                            Runtime
foo    f6717e   Up 2 seconds   3 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch stop foo
$ pouch ps -a
Name   ID       Status                 Created         Image                                            Runtime
foo    f6717e   Stopped (0) 1 minute   2 minutes ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch wait foo
0`
}
