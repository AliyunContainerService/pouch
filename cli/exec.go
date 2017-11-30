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
		Short: "Exec a process in a running container",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return e.runExec(args)
		},
	}
	e.addFlags()
}

// addFlags adds flags for specific command.
func (e *ExecCommand) addFlags() {
	flagSet := e.cmd.Flags()
	flagSet.BoolVarP(&e.Detach, "detach", "d", false, "Run the process in the background")
	flagSet.BoolVarP(&e.Terminal, "tty", "t", false, "Allocate a tty device")
	flagSet.BoolVarP(&e.Interactive, "interactive", "i", false, "Open container's STDIN")
}

// runExec is the entry of ExecCommand command.
func (e *ExecCommand) runExec(args []string) error {
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
		return fmt.Errorf("failed to create exec: %v", err)
	}

	// start exec process.
	startExecConfig := &types.ExecStartConfig{
		Detach: e.Detach,
		Tty:    e.Terminal && e.Interactive,
	}

	conn, reader, err := apiClient.ContainerStartExec(createResp.ID, startExecConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %v", err)
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
			return fmt.Errorf("failed to set raw mode")
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
	return nil
}
