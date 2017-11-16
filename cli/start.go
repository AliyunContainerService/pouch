package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// StartCommand use to implement 'start' command, it start a container.
type StartCommand struct {
	baseCommand
	attach bool
	stdin  bool
}

// Init initialize start command.
func (s *StartCommand) Init(c *Cli) {
	s.cli = c

	s.cmd = &cobra.Command{
		Use:   "start [container]",
		Short: "Start a created container",
		Args:  cobra.MinimumNArgs(1),
	}

	s.cmd.Flags().BoolVarP(&s.attach, "attach", "a", false, "attach the container's io or not")
	s.cmd.Flags().BoolVarP(&s.stdin, "interactive", "i", false, "attach container's stdin")
}

// Run is the entry of start command.
func (s *StartCommand) Run(args []string) {
	container := args[0]

	// attach to io.
	apiClient := s.cli.Client()

	var wait chan struct{}
	if s.attach || s.stdin {
		in, out, err := setRawMode(s.stdin, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to set raw mode")
			return
		}
		defer func() {
			if err := restoreMode(in, out); err != nil {
				fmt.Fprintf(os.Stderr, "failed to restore term mode")
			}
		}()

		conn, br, err := apiClient.ContainerAttach(container, s.stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to attach container: %v \n", err)
			return
		}

		wait = make(chan struct{})
		go func() {
			io.Copy(os.Stdout, br)
			close(wait)
		}()
		go func() {
			io.Copy(conn, os.Stdin)
			close(wait)
		}()
	}

	// start container
	if err := apiClient.ContainerStart(container, ""); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start container %s: %v\n", container, err)
		return
	}

	// wait the io to finish.
	if s.attach || s.stdin {
		<-wait
	}
}

func setRawMode(stdin, stdout bool) (*terminal.State, *terminal.State, error) {
	var (
		in  *terminal.State
		out *terminal.State
		err error
	)

	if stdin {
		if in, err = terminal.MakeRaw(0); err != nil {
			return nil, nil, err
		}
	}
	if stdout {
		if out, err = terminal.MakeRaw(1); err != nil {
			return nil, nil, err
		}
	}

	return in, out, nil
}

func restoreMode(in, out *terminal.State) error {
	if in != nil {
		if err := terminal.Restore(0, in); err != nil {
			return err
		}
	}
	if out != nil {
		if err := terminal.Restore(1, out); err != nil {
			return err
		}
	}
	return nil
}
