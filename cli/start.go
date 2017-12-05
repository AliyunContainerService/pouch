package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// startDescription is used to describe start command in detail and auto generate command doc.
// TODO: add description
var startDescription = ""

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
		Short: "Start a created or stopped container",
		Long:  startDescription,
		Args:  cobra.MinimumNArgs(1),
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
	flagSet.BoolVarP(&s.attach, "attach", "a", false, "Attach container's STDOUT and STDERR")
	flagSet.BoolVarP(&s.stdin, "interactive", "i", false, "Attach container's STDIN")
}

// runStart is the entry of start command.
func (s *StartCommand) runStart(args []string) error {
	container := args[0]

	// attach to io.
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

		conn, br, err := apiClient.ContainerAttach(container, s.stdin)
		if err != nil {
			return fmt.Errorf("failed to attach container: %v", err)
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
		return fmt.Errorf("failed to start container %s: %v", container, err)
	}

	// wait the io to finish.
	if s.attach || s.stdin {
		<-wait
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
// TODO: add example
func startExample() string {
	example := `# pouch start ${containerID} -a -i		
/ # ls /		
bin   dev   etc   home  proc  root  run   sys   tmp   usr   var		
/ # exit`

	return example
}
