package main

import (
	"fmt"
	"os"

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

	path := fmt.Sprintf("/containers/%s/start", container)

	req, err := s.cli.NewPostRequest(path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to new post request: %v \n", err)
		return
	}

	response := req.Send()
	if err := response.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to request: %v \n", err)
		return
	}
	defer response.Close()

	fmt.Printf("succeed in starting container: %s \n", container)
}
