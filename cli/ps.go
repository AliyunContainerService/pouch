package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/spf13/cobra"
)

// psDescription is used to describe ps command in detail and auto generate command doc.
var psDescription = "\nList Containers with container name, ID, status, image reference, runtime and creation time."

// containerList is used to save the container list.
type containerList []*types.Container

// PsCommand is used to implement 'ps' command.
type PsCommand struct {
	baseCommand
}

// Init initializes PsCommand command.
func (p *PsCommand) Init(c *Cli) {
	p.cli = c
	p.cmd = &cobra.Command{
		Use:   "ps",
		Short: "List all containers",
		Long:  psDescription,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runPs(args)
		},
		Example: psExample(),
	}
	p.addFlags()
}

// addFlags adds flags for specific command.
func (p *PsCommand) addFlags() {
	// TODO: add flags here
}

// runPs is the entry of PsCommand command.
func (p *PsCommand) runPs(args []string) error {
	apiClient := p.cli.Client()

	var containers containerList
	containers, err := apiClient.ContainerList()
	if err != nil {
		return fmt.Errorf("failed to get container list: %v", err)
	}

	sort.Sort(containers)

	display := p.cli.NewTableDisplay()
	display.AddRow([]string{"Name", "ID", "Status", "Image", "Runtime", "Created"})
	for _, c := range containers {
		created, err := utils.FormatCreatedTime(c.Created)
		if err != nil {
			return err
		}
		display.AddRow([]string{c.Names[0], c.ID[:6], c.Status, c.Image, c.HostConfig.Runtime, created})
	}
	display.Flush()
	return nil
}

// psExample shows examples in ps command, and is used in auto-generated cli docs.
func psExample() string {
	return `$ pouch ps
Name     ID       Status    Image                              Runtime   Created
1dad17   1dad17   stopped   docker.io/library/busybox:latest   runv      1 hour ago
505571   505571   stopped   docker.io/library/busybox:latest   runc      1 hour ago
`
}

// Len implements the sort interface.
func (c containerList) Len() int {
	return len(c)
}

// Swap implements the sort interface.
func (c containerList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// Less implements the sort interface.
func (c containerList) Less(i, j int) bool {
	ivalue := time.Unix(0, c[i].Created)
	jvalue := time.Unix(0, c[j].Created)
	return ivalue.After(jvalue)
}
