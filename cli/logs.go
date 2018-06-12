package main

import (
	"context"
	"io"
	"os"

	"github.com/alibaba/pouch/apis/types"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/spf13/cobra"
)

var validDrivers = map[string]bool{
	"json-file": true,
	"journald":  true,
}

// logsDescription is used to describe logs command in detail and auto generate command doc.
var logsDescription = ""

// LogsCommand use to implement 'logs' command, it is used to print a container's logs
type LogsCommand struct {
	baseCommand
	details    bool
	follow     bool
	since      string
	tail       string
	until      string
	timestamps bool
}

// Init initialize logs command.
func (lc *LogsCommand) Init(c *Cli) {
	lc.cli = c
	lc.cmd = &cobra.Command{
		Use:   "logs [OPTIONS] CONTAINER",
		Short: "Print a container's logs",
		Long:  logsDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return lc.runLogs(args)
		},
		Example: logsExample(),
	}
	lc.addFlags()
}

// addFlags adds flags for specific command.
func (lc *LogsCommand) addFlags() {
	flagSet := lc.cmd.Flags()
	flagSet.BoolVarP(&lc.follow, "follow", "f", false, "Follow log output")
	flagSet.StringVarP(&lc.since, "since", "", "", "Show logs since timestamp (e.g. 2013-01-02T13:23:37) or relative (e.g. 42m for 42 minutes)")
	flagSet.StringVarP(&lc.until, "until", "", "", "Show logs before timestamp (e.g. 2013-01-02T13:23:37) or relative (e.g. 42m for 42 minutes)")
	flagSet.StringVarP(&lc.tail, "tail", "", "all", "Number of lines to show from the end of the logs default \"all\"")
	flagSet.BoolVarP(&lc.timestamps, "timestamps", "t", false, "Show timestamps")

	// TODO(fuwei): support the detail functionality
}

// runLogs is the entry of LogsCommand command.
func (lc *LogsCommand) runLogs(args []string) error {
	containerName := args[0]

	ctx := context.Background()
	apiClient := lc.cli.Client()

	opts := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      lc.since,
		Until:      lc.until,
		Timestamps: lc.timestamps,
		Follow:     lc.follow,
		Tail:       lc.tail,
	}

	body, err := apiClient.ContainerLogs(ctx, containerName, opts)
	if err != nil {
		return err
	}

	defer body.Close()

	c, err := apiClient.ContainerGet(ctx, containerName)
	if err != nil {
		return err
	}

	if c.Config.Tty {
		_, err = io.Copy(os.Stdout, body)
	} else {
		_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, body)
	}
	return err
}

// logsExample shows examples in logs command, and is used in auto-generated cli docs.
func logsExample() string {
	return ``
}
