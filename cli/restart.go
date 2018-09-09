package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// restartDescription is used to describe restart command in detail and auto generate command doc.
var restartDescription = "restart one or more containers"

// RestartCommand uses to implement 'restart' command, it restarts one or more containers.
type RestartCommand struct {
	baseCommand
	timeout int
}

// Init initialize restart command.
func (rc *RestartCommand) Init(c *Cli) {
	rc.cli = c

	rc.cmd = &cobra.Command{
		Use:   "restart [OPTION] CONTAINER [CONTAINER...]",
		Short: "restart one or more containers",
		Long:  restartDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return rc.runRestart(args)
		},
		Example: restartExample(),
	}
	rc.addFlags()
}

// addFlags adds flags for specific command.
func (rc *RestartCommand) addFlags() {
	flagSet := rc.cmd.Flags()
	flagSet.IntVarP(&rc.timeout, "time", "t", 10, "Seconds to wait for stop before killing the container")
}

// runRestart is the entry of restart command.
func (rc *RestartCommand) runRestart(args []string) error {
	ctx := context.Background()
	apiClient := rc.cli.Client()

	var errs []string
	for _, name := range args {
		if err := apiClient.ContainerRestart(ctx, name, strconv.Itoa(rc.timeout)); err != nil {
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

// restartExample shows examples in restart command, and is used in auto-generated cli docs.
func restartExample() string {
	return `$ pouch ps -a
Name     ID       Status    Image                              Runtime
foo      71b9c1   Stopped   docker.io/library/busybox:latest   runc
$ pouch restart foo
$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Running   docker.io/library/busybox:latest   runc`
}
