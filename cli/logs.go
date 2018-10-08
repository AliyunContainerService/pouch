package main

import (
	"context"
	"io"
	"os"

	"github.com/alibaba/pouch/apis/types"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/spf13/cobra"
)

// logsDescription is used to describe logs command in detail and auto generate command doc.
var logsDescription = "Get container's logs"

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
	return `$ pouch ps 
Name     ID       Status      Created      Image                                          Runtime
073f29   073f29   Up 1 day    2 days ago   registry.hub.docker.com/library/redis:latest   runc
$ pouch logs 073f29
1:C 04 Sep 05:42:01.600 # oO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo
1:C 04 Sep 05:42:01.601 # Redis version=4.0.11, bits=64, commit=00000000, modified=0, pid=1, just started
1:C 04 Sep 05:42:01.601 # Warning: no config file specified, using the default config. In order to specify a config file use redis-server /path/to/redis.conf
1:M 04 Sep 05:42:01.601 * Increased maximum number of open files to 10032 (it was originally set to 1024).
1:M 04 Sep 05:42:01.602 * Running mode=standalone, port=6379.
1:M 04 Sep 05:42:01.602 # WARNING: The TCP backlog setting of 511 cannot be enforced because /proc/sys/net/core/somaxconn is set to the lower value of 128.
1:M 04 Sep 05:42:01.602 # Server initialized
1:M 04 Sep 05:42:01.602 # WARNING overcommit_memory is set to 0! Background save may fail under low memory condition. To fix this issue add 'vm.overcommit_memory = 1' to /etc/sysctl.conf and then reboot or run the command 'sysctl vm.overcommit_memory=1' for this to take effect.
1:M 04 Sep 05:42:01.602 # WARNING you have Transparent Huge Pages (THP) support enabled in your kernel. This will create latency and memory usage issues with Redis. To fix this issue run the command 'echo never > /sys/kernel/mm/transparent_hugepage/enabled' as root, and add it to your /etc/rc.local in order to retain the setting after a reboot. Redis must be restarted after THP is disabled.
1:M 04 Sep 05:42:01.602 * Ready to accept connections`
}
