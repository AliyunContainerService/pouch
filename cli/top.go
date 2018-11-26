package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// topDescription
var topDescription = "top command is to display the running processes of a container. " +
	"You can add options just like using Linux ps command."

// TopCommand use to implement 'top' command, it displays all processes in a container.
type TopCommand struct {
	baseCommand
}

// Init initialize top command.
func (top *TopCommand) Init(c *Cli) {
	top.cli = c
	top.cmd = &cobra.Command{
		Use:   "top CONTAINER [ps OPTIONS]",
		Short: "Display the running processes of a container",
		Long:  topDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return top.runTop(args)
		},
		Example: topExamples(),
	}
	top.addFlags()
}

// addFlags adds flags for specific command.
func (top *TopCommand) addFlags() {
	flagSet := top.cmd.Flags()
	flagSet.SetInterspersed(false)
}

// runTop is the entry of top command.
func (top *TopCommand) runTop(args []string) error {
	ctx := context.Background()
	apiClient := top.cli.Client()

	container := args[0]

	arguments := args[1:]

	procList, err := apiClient.ContainerTop(ctx, container, arguments)
	if err != nil {
		return fmt.Errorf("failed to execute top command in container %s: %v", container, err)
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 8, 4, ' ', 0)
	fmt.Fprintln(w, strings.Join(procList.Titles, "\t"))

	for _, ps := range procList.Processes {
		fmt.Fprintln(w, strings.Join(ps, "\t"))
	}
	w.Flush()
	return nil
}

// topExamples shows examples in top command, and is used in auto-generated cli docs.
func topExamples() string {
	return `$ pouch top 44f675
UID     PID      PPID     C    STIME    TTY    TIME        CMD
root    28725    28714    0    3月14     ?      00:00:00    sh
`
}
