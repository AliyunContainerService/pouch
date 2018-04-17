package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// startDescription is used to describe start command in detail and auto generate command doc.
var startDescription = "Start a created container object in Pouchd. " +
	"When starting, the relevant resource preserved during creating period comes into use." +
	"This is useful when you wish to start a container which has been created in advance." +
	"The container you started will be running if no error occurs."

// StartCommand use to implement 'start' command, it start a container.
type StartCommand struct {
	baseCommand
	detachKeys string
	attach     bool
	stdin      bool
}

// Init initialize start command.
func (s *StartCommand) Init(c *Cli) {
	s.cli = c
	s.cmd = &cobra.Command{
		Use:   "start [OPTIONS] CONTAINER",
		Short: "Start a created or stopped container",
		Long:  startDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.runStart(args)
		},
		Example: startExample(),
	}
	s.addFlags()
}

// addFlags adds flags for specific command.
func (s *StartCommand) addFlags() {
	flagSet := s.cmd.Flags()
	flagSet.StringVar(&s.detachKeys, "detach-keys", "", "Override the key sequence for detaching a container")
	flagSet.BoolVarP(&s.attach, "attach", "a", false, "Attach container's STDOUT and STDERR")
	flagSet.BoolVarP(&s.stdin, "interactive", "i", false, "Attach container's STDIN")
}

// runStart is the entry of start command.
func (s *StartCommand) runStart(args []string) error {
	container := args[0]

	// attach to io.
	ctx := context.Background()
	apiClient := s.cli.Client()

	var wait chan struct{}
	if s.attach || s.stdin {
		in, out, err := setRawMode(s.stdin, false)
		if err != nil {
			return fmt.Errorf("failed to set raw mode")
		}
		defer func() {
			if err := restoreMode(in, out); err != nil {
				fmt.Fprintf(os.Stderr, "failed to restore term mode")
			}
		}()

		conn, br, err := apiClient.ContainerAttach(ctx, container, s.stdin)
		if err != nil {
			return fmt.Errorf("failed to attach container: %v", err)
		}
		defer conn.Close()

		wait = make(chan struct{})
		go func() {
			io.Copy(os.Stdout, br)
			close(wait)
		}()
		go func() {
			io.Copy(conn, os.Stdin)
		}()
	}

	// start container
	if err := apiClient.ContainerStart(ctx, container, s.detachKeys); err != nil {
		return fmt.Errorf("failed to start container %s: %v", container, err)
	}

	// wait the io to finish.
	if s.attach || s.stdin {
		<-wait
	}

	info, err := apiClient.ContainerGet(ctx, container)
	if err != nil {
		return err
	}

	code := info.State.ExitCode
	if code != 0 {
		return ExitError{Code: int(code)}
	}

	return nil
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

// startExample shows examples in start command, and is used in auto-generated cli docs.
func startExample() string {
	return `$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Created   docker.io/library/busybox:latest   runc
$ pouch start foo
$ pouch ps
Name     ID       Status    Image                              Runtime
foo      71b9c1   Running   docker.io/library/busybox:latest   runc`
}
