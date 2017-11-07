package main

import (
	"fmt"
	"os"

	"github.com/alibaba/pouch/client"

	"github.com/spf13/cobra"
)

// StartCommand use to implement 'start' command, it start a container.
type StartCommand struct {
	baseCommand
}

// Init initialize start command.
func (s *StartCommand) Init(c *Cli) {
	s.cli = c

	s.cmd = &cobra.Command{
		Use:   "start [container]",
		Short: "Start a created container",
		Args:  cobra.MinimumNArgs(1),
	}
}

// Run is the entry of start command.
func (s *StartCommand) Run(args []string) {
	container := args[0]

	client, err := client.New("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to new client: %v\n", err)
		return
	}

	err = client.ContainerStart(container, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start container %s: %v\n", container, err)
		return
	}

	fmt.Printf("succeed in starting container: %s \n", container)
}
