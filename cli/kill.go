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

// KillCommand use to implement 'kill' command, it kills a container.
type KillCommand struct {
	baseCommand
	signal string
}

// Init initialize kill command.
func (kc *KillCommand) Init(c *Cli) {
	kc.cli = c
	kc.cmd = &cobra.Command{
		Use:   "kill [OPTIONS] CONTAINER [CONTAINER...]",
		Short: "kill one or more running containers",
		Long:  killDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return kc.runKill(args)
		},
		Example: killExample(),
	}
	kc.addFlags()
}

// addFlags adds flags for specific command.
func (kc *KillCommand) addFlags() {
	flagSet := kc.cmd.Flags()
	flagSet.StringVarP(&kc.signal, "signal", "s", "KILL", "Signal to send to the container (default \"KILL\")")
}

func (kc *KillCommand) runKill(args []string) error {
	ctx := context.Background()
	apiClient := kc.cli.Client()

	var errs []string
	for _, name := range args {
		if err := apiClient.ContainerKill(ctx, name, kc.signal); err != nil {
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

// killExample shows examples in kill command, and is used in auto-generated documentation.
func killExample() string {
	return `$ pouch ps -a
Name   ID       Status    		Created         Image                                            Runtime
foo2   5a0ede   Up 2 seconds   	3 second ago    registry.hub.docker.com/library/busybox:latest   runc
foo1   e05637   Up 6 seconds   	7 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch kill foo1
foo1
$ pouch ps
Name   ID       Status         Created          Image                                            Runtime
foo2   5a0ede   Up 11 seconds   12 seconds ago   registry.hub.docker.com/library/busybox:latest   runc`
}
