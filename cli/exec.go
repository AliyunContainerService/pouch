package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/alibaba/pouch/apis/types"

	"github.com/spf13/cobra"
)

// ExecCommand is used to implement 'exec' command.
type ExecCommand struct {
	baseCommand
	Interactive bool
	Terminal    bool
	Detach      bool
}

// Init initializes ExecCommand command.
func (e *ExecCommand) Init(c *Cli) {
	e.cli = c
	e.cmd = &cobra.Command{
		Use:   "exec [container]",
		Short: "exec a process in container",
		Args:  cobra.MinimumNArgs(2),
	}

	e.cmd.Flags().BoolVarP(&e.Detach, "detach", "d", false, "run process in the backgroud")
	e.cmd.Flags().BoolVarP(&e.Terminal, "tty", "t", false, "allocate a tty device")
	e.cmd.Flags().BoolVarP(&e.Interactive, "interactive", "i", false, "open container's stdin io")
}

// Run is the entry of ExecCommand command.
func (e *ExecCommand) Run(args []string) {
	apiClient := e.cli.Client()

	// create exec process.
	id := args[0]
	command := args[1]

	createExecConfig := &types.ExecCreateConfig{
		Cmd:          strings.Fields(command),
		Tty:          e.Terminal,
		Detach:       e.Detach,
		AttachStderr: !e.Detach,
		AttachStdout: !e.Detach,
		AttachStdin:  !e.Detach && e.Interactive,
		Privileged:   false,
		User:         "",
	}

	createResp, err := apiClient.ContainerCreateExec(id, createExecConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create exec: %v", err)
		return
	}

	// start exec process.
	startExecConfig := &types.ExecStartConfig{
		Detach: e.Detach,
		Tty:    e.Terminal && e.Interactive,
	}

	conn, reader, err := apiClient.ContainerStartExec(createResp.ID, startExecConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create exec: %v", err)
		return
	}

	// handle stdio.
	var wg sync.WaitGroup

	if createExecConfig.AttachStderr || createExecConfig.AttachStdout {
		wg.Add(1)
		go func() {
			defer wg.Done()
			io.Copy(os.Stdout, reader)
		}()
	}
	if createExecConfig.AttachStdin {
		in, out, err := setRawMode(true, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to set raw mode")
			return
		}
		defer func() {
			if err := restoreMode(in, out); err != nil {
				fmt.Fprintf(os.Stderr, "failed to restore term mode")
			}
		}()

		go func() {
			io.Copy(conn, os.Stdin)
		}()
	}

	wg.Wait()
}
