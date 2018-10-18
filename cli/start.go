package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alibaba/pouch/apis/types"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

// startDescription is used to describe start command in detail and auto generate command doc.
var startDescription = "Start one or more created container objects in Pouchd. " +
	"When starting, the relevant resource preserved during creating period comes into use." +
	"This is useful when you wish to start a container which has been created in advance." +
	"The container you started will be running if no error occurs."

// StartCommand use to implement 'start' command, it start one or more containers.
type StartCommand struct {
	baseCommand
	detachKeys string
	attach     bool
	stdin      bool
	checkpoint string
	cpDir      string
}

// Init initialize start command.
func (s *StartCommand) Init(c *Cli) {
	s.cli = c
	s.cmd = &cobra.Command{
		Use:   "start [OPTIONS] CONTAINER [CONTAINER...]",
		Short: "Start one or more created or stopped containers",
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
	flagSet.StringVar(&s.detachKeys, "detach-keys", "", "Override the key sequence for detaching a container")
	flagSet.BoolVarP(&s.attach, "attach", "a", false, "Attach container's STDOUT and STDERR")
	flagSet.BoolVarP(&s.stdin, "interactive", "i", false, "Attach container's STDIN")
	flagSet.StringVar(&s.checkpoint, "checkpoint", "", "Restore container state from the checkpoint")
	flagSet.StringVar(&s.cpDir, "checkpoint-dir", "", "Directory to store checkpoints images")
}

// runStart is the entry of start command.
func (s *StartCommand) runStart(args []string) error {
	ctx := context.Background()
	apiClient := s.cli.Client()

	// attach to io.
	if s.attach || s.stdin {
		var wait chan struct{}
		// If we want to attach to a container, we should make sure we only have one container.
		if len(args) > 1 {
			return fmt.Errorf("cannot start and attach multiple containers at once")
		}

		in, out, err := setRawMode(s.stdin, false)
		if err != nil {
			return fmt.Errorf("failed to set raw mode")
		}
		defer func() {
			if err := restoreMode(in, out); err != nil {
				fmt.Fprintf(os.Stderr, "failed to restore term mode")
			}
		}()

		container := args[0]
		c, err := apiClient.ContainerGet(ctx, container)
		if err != nil {
			return err
		}

		conn, br, err := apiClient.ContainerAttach(ctx, container, s.stdin)
		if err != nil {
			return fmt.Errorf("failed to attach container: %v", err)
		}
		defer conn.Close()

		wait = make(chan struct{})
		go func() {
			if !c.Config.Tty {
				stdcopy.StdCopy(os.Stdout, os.Stderr, br)
			} else {
				io.Copy(os.Stdout, br)
			}
			close(wait)
		}()
		go func() {
			io.Copy(conn, os.Stdin)
		}()

		// start container
		if err := apiClient.ContainerStart(ctx, container, types.ContainerStartOptions{
			DetachKeys:    s.detachKeys,
			CheckpointID:  s.checkpoint,
			CheckpointDir: s.cpDir,
		}); err != nil {
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
	} else {
		// We're not going to attach to any container, so we just start as many containers as we want.
		var errs []string
		for _, name := range args {
			if err := apiClient.ContainerStart(ctx, name, types.ContainerStartOptions{
				DetachKeys:    s.detachKeys,
				CheckpointID:  s.checkpoint,
				CheckpointDir: s.cpDir,
			}); err != nil {
				errs = append(errs, err.Error())
				continue
			}
			fmt.Printf("%s\n", name)
		}

		if len(errs) > 0 {
			return errors.New("failed to start containers: " + strings.Join(errs, ""))
		}
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
	return `$ pouch ps -a
Name   ID       Status    Created         Image                                            Runtime
foo2   5a0ede   created   1 second ago    registry.hub.docker.com/library/busybox:latest   runc
foo1   e05637   created   6 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
$ pouch start foo1 foo2
foo1
foo2
$ pouch ps
Name   ID       Status         Created          Image                                            Runtime
foo2   5a0ede   Up 2 seconds   12 seconds ago   registry.hub.docker.com/library/busybox:latest   runc
foo1   e05637   Up 3 seconds   17 seconds ago   registry.hub.docker.com/library/busybox:latest   runc`
}
