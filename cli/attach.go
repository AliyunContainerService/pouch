package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/client"
	"github.com/alibaba/pouch/pkg/ioutils"

	"github.com/docker/docker/pkg/term"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// AttachDescription is used to describe attach command in detail and auto generate command doc.
var AttachDescription = "Attach local standard input, output, and error streams to a running container"

var defaultEscapeKeys = []byte{16, 17}

// AttachCommand is used to implement 'attach' command.
type AttachCommand struct {
	baseCommand

	// flags for attach command
	noStdin    bool
	detachKeys string
}

// Init initialize "attach" command.
func (ac *AttachCommand) Init(c *Cli) {
	ac.cli = c
	ac.cmd = &cobra.Command{
		Use:   "attach [OPTIONS] CONTAINER",
		Short: "Attach local standard input, output, and error streams to a running container",
		Long:  AttachDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ac.runAttach(args)
		},
		Example: ac.example(),
	}
	ac.addFlags()
}

// addFlags adds flags for specific command.
func (ac *AttachCommand) addFlags() {
	flagSet := ac.cmd.Flags()
	flagSet.BoolVar(&ac.noStdin, "no-stdin", false, "Do not attach STDIN")
	flagSet.StringVar(&ac.detachKeys, "detach-keys", "", "Override the key sequence for detaching a container")
	// TODO: sig-proxy will be supported in the future.
	//flagSet.BoolVar(&ac.sigProxy, "sig-proxy", true, "Proxy all received signals to the process")
}

func inspectAndCheckState(ctx context.Context, cli client.CommonAPIClient, name string) (*types.ContainerJSON, error) {
	c, err := cli.ContainerGet(ctx, name)
	if err != nil {
		return nil, err
	}
	if !c.State.Running {
		return nil, errors.New("You cannot attach to a stopped container, start it first")
	}
	if c.State.Paused {
		return nil, errors.New("You cannot attach to a paused container, unpause it first")
	}
	if c.State.Restarting {
		return nil, errors.New("You cannot attach to a restarting container, wait until it is running")
	}

	return c, nil
}

// runAttach is used to attach a container.
func (ac *AttachCommand) runAttach(args []string) error {
	name := args[0]

	ctx := context.Background()
	apiClient := ac.cli.Client()

	c, err := inspectAndCheckState(ctx, apiClient, name)
	if err != nil {
		return err
	}

	if err := checkTty(!ac.noStdin, c.Config.Tty, os.Stdin.Fd()); err != nil {
		return err
	}

	var inReader io.Reader = os.Stdin
	if !ac.noStdin && c.Config.Tty {
		in, out, err := setRawMode(!ac.noStdin, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to set raw mode %s", err)
			return fmt.Errorf("failed to set raw mode")
		}
		defer func() {
			if err := restoreMode(in, out); err != nil {
				fmt.Fprintf(os.Stderr, "failed to restore term mode %s", err)
			}
		}()

		escapeKeys := defaultEscapeKeys
		// Wrap the input to detect detach escape sequence.
		// Use default escape keys if an invalid sequence is given.
		if ac.detachKeys != "" {
			customEscapeKeys, err := term.ToBytes(ac.detachKeys)
			if err != nil {
				return fmt.Errorf("invalid detach keys (%s) provided", ac.detachKeys)
			}
			escapeKeys = customEscapeKeys
		}
		inReader = ioutils.NewReadCloserWrapper(term.NewEscapeProxy(os.Stdin, escapeKeys), os.Stdin.Close)
	}

	conn, br, err := apiClient.ContainerAttach(ctx, name, !ac.noStdin)
	if err != nil {
		return fmt.Errorf("failed to attach container: %v", err)
	}
	defer conn.Close()

	outputDone := make(chan error, 1)
	go func() {
		var err error
		_, err = io.Copy(os.Stdout, br)
		if err != nil {
			logrus.Debugf("Error receive stdout: %s", err)
		}
		outputDone <- err
	}()

	inputDone := make(chan struct{})
	detached := make(chan error, 1)
	go func() {
		if !ac.noStdin {
			_, err := io.Copy(conn, inReader)
			// close write if receive CTRL-D
			if cw, ok := conn.(ioutils.CloseWriter); ok {
				cw.CloseWrite()
			}
			if _, ok := err.(term.EscapeError); ok {
				detached <- err
			}
			if err != nil {
				logrus.Debugf("Error send stdin: %s", err)
			}
		}
		close(inputDone)

	}()

	select {
	case err := <-outputDone:
		if err != nil {
			logrus.Debugf("receive stdout error: %s", err)
			return err
		}
	case <-inputDone:
		select {
		// Wait for output to complete streaming.
		case err := <-outputDone:
			logrus.Debugf("receive stdout error: %s", err)
			return err
		case <-ctx.Done():
		}
	case err := <-detached:
		// Got a detach key sequence.
		return err
	case <-ctx.Done():
	}

	return nil
}

// example shows examples in attach command, and is used in auto-generated cli docs.
func (ac *AttachCommand) example() string {
	return `$ pouch run -d --name foo busybox sh -c 'while true; do sleep 1; echo hello; done'
Name     ID       Status    Image                              Runtime
foo      71b9c1   Running   docker.io/library/busybox:latest   runc
$ pouch attach foo
hello
hello
hello`
}
