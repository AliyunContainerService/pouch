package main

import (
	"context"
	"fmt"

	//"github.com/alibaba/pouch/apis/types"
	//"github.com/alibaba/pouch/pkg/reference"

	"github.com/spf13/cobra"
)

// topDescription
var topDescription = ""

// TopCommand use to implement 'top' command, it displays all processes in a container.
type TopCommand struct {
	baseCommand
	args []string
}

// Init initialize top command.
func (top *TopCommand) Init(c *Cli) {
	top.cli = c
	top.cmd = &cobra.Command{
		Use:   "top CONTAINER",
		Short: "Display the running processes of a container",
		Long:  topDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return top.runTop(args)
		},
		Example: topExamples(),
	}
}

// runTop is the entry of top command.
func (top *TopCommand) runTop(args []string) error {
	ctx := context.Background()
	apiClient := top.cli.Client()

	container := args[0]

	arguments := args[1:]

	resp, err := apiClient.ContainerTop(ctx, container, arguments)
	if err != nil {
		return fmt.Errorf("failed to execute top command in container %s: %v", container, err)
	}

	fmt.Println(resp)
	return nil
}

// topExamples shows examples in top command, and is used in auto-generated cli docs.
func topExamples() string {
	return ``
}
