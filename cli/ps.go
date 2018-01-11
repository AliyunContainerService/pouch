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
var psDescription = "\nList Containers with container name, ID, status, creation time, image reference and runtime."

// containerList is used to save the container list.
type containerList []*types.Container

// PsCommand is used to implement 'ps' command.
type PsCommand struct {
	baseCommand

	// flags for ps command
	flagAll   bool
	flagQuiet bool
}

// Init initializes PsCommand command.
func (p *PsCommand) Init(c *Cli) {
	p.cli = c
	p.cmd = &cobra.Command{
		Use:   "ps [OPTIONS]",
		Short: "List containers",
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
	flagSet := p.cmd.Flags()
	flagSet.BoolVarP(&p.flagAll, "all", "a", false, "Show all containers (default shows just running)")
	flagSet.BoolVarP(&p.flagQuiet, "quiet", "q", false, "Only show numeric IDs")
}

// runPs is the entry of PsCommand command.
func (p *PsCommand) runPs(args []string) error {
	apiClient := p.cli.Client()

	var containers containerList
	containers, err := apiClient.ContainerList(p.flagAll)
	if err != nil {
		return fmt.Errorf("failed to get container list: %v", err)
	}

	sort.Sort(containers)

	if p.flagQuiet {
		for _, c := range containers {
			fmt.Println(c.ID[:6])
		}
		return nil
	}

	display := p.cli.NewTableDisplay()
	display.AddRow([]string{"Name", "ID", "Status", "Created", "Image", "Runtime"})

	for _, c := range containers {
		created, err := utils.FormatTimeInterval(c.Created)
		if err != nil {
			return err
		}

		display.AddRow([]string{c.Names[0], c.ID[:6], c.Status, created + " ago", c.Image, c.HostConfig.Runtime})
	}

	display.Flush()
	return nil
}

// psExample shows examples in ps command, and is used in auto-generated cli docs.
func psExample() string {
	return `$ pouch ps
Name   ID       Status          Created          Image                              Runtime
2      e42c68   Up 15 minutes   16 minutes ago   docker.io/library/busybox:latest   runc
1      a8c2ea   Up 16 minutes   17 minutes ago   docker.io/library/busybox:latest   runc

$ pouch ps -a
Name   ID       Status          Created          Image                              Runtime
3      faf132   created         16 seconds ago   docker.io/library/busybox:latest   runc
2      e42c68   Up 16 minutes   16 minutes ago   docker.io/library/busybox:latest   runc
1      a8c2ea   Up 17 minutes   18 minutes ago   docker.io/library/busybox:latest   runc

$ pouch ps -q
e42c68
a8c2ea

$ pouch ps -a -q
faf132
e42c68
a8c2ea
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
	iValue := time.Unix(0, c[i].Created)
	jValue := time.Unix(0, c[j].Created)
	return iValue.After(jValue)
}
