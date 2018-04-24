package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/alibaba/pouch/apis/types"

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
	flagSet.BoolVarP(&lc.details, "details", "", false, "Show extra provided to logs")
	flagSet.BoolVarP(&lc.follow, "follow", "f", false, "Follow log output")
	flagSet.StringVarP(&lc.since, "since", "", "", "Show logs since timestamp")
	flagSet.StringVarP(&lc.tail, "tail", "", "all", "Number of lines to show from the end of the logs default \"all\"")
	flagSet.BoolVarP(&lc.timestamps, "timestamps", "t", false, "Show timestamps")
}

// runLogs is the entry of LogsCommand command.
func (lc *LogsCommand) runLogs(args []string) error {
	// TODO

	containerName := args[0]

	ctx := context.Background()
	apiClient := lc.cli.Client()

	opts := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      lc.since,
		Timestamps: lc.timestamps,
		Follow:     lc.follow,
		Tail:       lc.tail,
		Details:    lc.details,
	}

	resp, err := apiClient.ContainerLogs(ctx, containerName, opts)
	if err != nil {
		return err
	}

	defer resp.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp)

	fmt.Printf(buf.String())

	return nil
}

// logsExample shows examples in logs command, and is used in auto-generated cli docs.
func logsExample() string {
	return ``
}
