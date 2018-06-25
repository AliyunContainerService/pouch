package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/alibaba/pouch/apis/types"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/spf13/cobra"
)

// execDescription is used to describe exec command in detail and auto generate command doc.
// TODO: add description
var execDescription = ""

// ExecCommand is used to implement 'exec' command.
type ExecCommand struct {
	baseCommand
	Interactive bool
	Terminal    bool
	Detach      bool
	User        string
}

// Init initializes ExecCommand command.
func (e *ExecCommand) Init(c *Cli) {
	e.cli = c
	e.cmd = &cobra.Command{
		Use:   "exec [OPTIONS] CONTAINER COMMAND [ARG...]",
		Short: "Exec a process in a running container",
		Long:  execDescription,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return e.runExec(args)
		},
		Example: execExample(),
	}
	e.addFlags()
}

// addFlags adds flags for specific command.
func (e *ExecCommand) addFlags() {
	flagSet := e.cmd.Flags()
	flagSet.SetInterspersed(false)
	flagSet.BoolVarP(&e.Detach, "detach", "d", false, "Run the process in the background")
	flagSet.BoolVarP(&e.Terminal, "tty", "t", false, "Allocate a tty device")
	// NOTE: set default interactive equal true for taking over running moby 1.12.6 container,
	// since this version always need container open stdin/stdout/stderr fifo.
	flagSet.BoolVarP(&e.Interactive, "interactive", "i", true, "Open container's STDIN")
	flagSet.StringVarP(&e.User, "user", "u", "", "Username or UID (format: <name|uid>[:<group|gid>])")
}

// runExec is the entry of ExecCommand command.
func (e *ExecCommand) runExec(args []string) error {
	ctx := context.Background()
	apiClient := e.cli.Client()

	// create exec process.
	id := args[0]
	command := args[1:]

	createExecConfig := &types.ExecCreateConfig{
		Cmd:          command,
		Tty:          e.Terminal,
		Detach:       e.Detach,
		AttachStderr: !e.Detach,
		AttachStdout: !e.Detach,
		AttachStdin:  !e.Detach && e.Interactive,
		Privileged:   false,
		User:         e.User,
	}

	createResp, err := apiClient.ContainerCreateExec(ctx, id, createExecConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %v", err)
	}

	// start exec process.
	startExecConfig := &types.ExecStartConfig{
		Detach: e.Detach,
		Tty:    e.Terminal,
	}

	conn, reader, err := apiClient.ContainerStartExec(ctx, createResp.ID, startExecConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %v", err)
	}
	defer conn.Close()

	// handle stdio.
	var wg sync.WaitGroup

	if createExecConfig.AttachStderr || createExecConfig.AttachStdout {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !e.Terminal {
				stdcopy.StdCopy(os.Stdout, os.Stderr, reader)
			} else {
				io.Copy(os.Stdout, reader)
			}
		}()
	}

	if createExecConfig.AttachStdin {
		if e.Terminal {
			in, out, err := setRawMode(true, false)
			if err != nil {
				return fmt.Errorf("failed to set raw mode")
			}
			defer func() {
				if err := restoreMode(in, out); err != nil {
					fmt.Fprintf(os.Stderr, "failed to restore term mode")
				}
			}()
		}

		go func() {
			io.Copy(conn, os.Stdin)
		}()
	}

	wg.Wait()

	execInfo, err := apiClient.ContainerExecInspect(ctx, createResp.ID)
	if err != nil {
		return err
	}

	code := execInfo.ExitCode
	if code != 0 {
		return ExitError{Code: int(code)}
	}

	return nil
}

// execExample shows examples in exec command, and is used in auto-generated cli docs.
func execExample() string {
	return `$ pouch exec -it 25bf50 ps
PID   USER     TIME  COMMAND
    1 root      0:00 /bin/sh
   38 root      0:00 ps
`
}
