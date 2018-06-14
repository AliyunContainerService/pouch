package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// killDescription is used to describe kill command in detail and auto generate command doc.
var killDescription = "Kill one or more running containers, the container will receive SIGKILL by default, " +
	"or the signal which is specified with the --signal option."

// KillCommand use to implement 'kill' command
type KillCommand struct {
	baseCommand
	signal string
}

// Init initialize kill command.
func (kill *KillCommand) Init(c *Cli) {
	kill.cli = c
	kill.cmd = &cobra.Command{
		Use:   "kill [OPTIONS] CONTAINER [CONTAINER...]",
		Short: "Kill one or more running containers",
		Long:  killDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return kill.runKill(args)
		},
		Example: killExample(),
	}
	kill.addFlags()
}

// addFlags adds flags for specific command.
func (kill *KillCommand) addFlags() {
	flagSet := kill.cmd.Flags()
	flagSet.StringVarP(&kill.signal, "signal", "s", "SIGKILL", "Signal to send to the container")
}

// runKill is the entry of kill command.
func (kill *KillCommand) runKill(args []string) error {
	ctx := context.Background()
	apiClient := kill.cli.Client()

	var errs []string
	for _, name := range args {
		if err := apiClient.ContainerKill(ctx, name, kill.signal); err != nil {
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
	return `$ pouch ps
Name            ID       Status          Created          Image                                            Runtime
foo             c926cf   Up 5 seconds    6 seconds ago    registry.hub.docker.com/library/busybox:latest   runc
$ pouch kill foo
foo
$ pouch ps -a
Name            ID       Status                     Created          Image                                            Runtime
foo             c926cf   Exited (137) 9 seconds     25 seconds ago   registry.hub.docker.com/library/busybox:latest   runc`
}
