package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/spf13/cobra"
)

// psDescription is used to describe ps command in detail and auto generate command doc.
var psDescription = "\nList Containers with container name, ID, status, creation time, image reference, runtime and quoted command."

// containerList is used to save the container list.
type containerList []*types.Container

// PsCommand is used to implement 'ps' command.
type PsCommand struct {
	baseCommand

	// flags for ps command
	flagAll     bool
	flagQuiet   bool
	flagNoTrunc bool
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
	flagSet.BoolVar(&p.flagNoTrunc, "no-trunc", false, "Do not truncate output")
}

// runPs is the entry of PsCommand command.
func (p *PsCommand) runPs(args []string) error {
	ctx := context.Background()
	apiClient := p.cli.Client()

	var containers containerList
	containers, err := apiClient.ContainerList(ctx, p.flagAll)
	if err != nil {
		return fmt.Errorf("failed to get container list: %v", err)
	}

	sort.Sort(containers)

	if p.flagQuiet {
		for _, c := range containers {
			id := c.ID[:6]
			if p.flagNoTrunc {
				id = c.ID
			}
			fmt.Println(id)
		}
		return nil
	}

	display := p.cli.NewTableDisplay()
	display.AddRow([]string{"Name", "ID", "Status", "Created", "Image", "Runtime", "Command"})

	for _, c := range containers {
		created, err := utils.FormatTimeInterval(c.Created)
		if err != nil {
			return err
		}

		id := c.ID[:6]
		if p.flagNoTrunc {
			id = c.ID
		}

		var command string
		if len(c.Command) > 20 {
			command = string([]byte(c.Command)[:17]) + "..."
		} else {
			command = c.Command
		}
		command = strings.Join([]string{`"`, command, `"`}, "")

		display.AddRow([]string{c.Names[0], id, c.Status, created + " ago", c.Image, c.HostConfig.Runtime, command})
	}
	display.Flush()
	return nil
}

// psExample shows examples in ps command, and is used in auto-generated cli docs.
func psExample() string {
	return `$ pouch ps
Name   ID       Status          Created          Image                              Runtime   Command
2      e42c68   Up 15 minutes   16 minutes ago   docker.io/library/busybox:latest   runc      "sh"
1      a8c2ea   Up 16 minutes   17 minutes ago   docker.io/library/busybox:latest   runc      "sh"

$ pouch ps -a
Name   ID       Status          Created          Image                              Runtime   Command
3      faf132   created         16 seconds ago   docker.io/library/busybox:latest   runc      "sh"
2      e42c68   Up 16 minutes   16 minutes ago   docker.io/library/busybox:latest   runc      "sh"
1      a8c2ea   Up 17 minutes   18 minutes ago   docker.io/library/busybox:latest   runc      "sh"

$ pouch ps -q
e42c68
a8c2ea

$ pouch ps -a -q
faf132
e42c68
a8c2ea

$ pouch ps --no-trunc
Name   ID                                                                 Status        Created        Image                            Runtime   Command
foo2   692c77587b38f60bbd91d986ec3703848d72aea5030e320d4988eb02aa3f9d48   Up 1 minute   1 minute ago   docker.io/library/redis:alpine   runc      "sh"
foo    18592900006405ee64788bd108ef1de3d24dc3add73725891f4787d0f8e036f5   Up 1 minute   1 minute ago   docker.io/library/redis:alpine   runc      "sh"

$ pouch ps --no-trunc -q
692c77587b38f60bbd91d986ec3703848d72aea5030e320d4988eb02aa3f9d48
18592900006405ee64788bd108ef1de3d24dc3add73725891f4787d0f8e036f5

$ pouch ps --no-trunc -a
Name   ID                                                                 Status         Created         Image                            Runtime   Command
foo3   63fd6371f3d614bb1ecad2780972d5975ca1ab534ec280c5f7d8f4c7b2e9989d   created        2 minutes ago   docker.io/library/redis:alpine   runc      "sh"
foo2   692c77587b38f60bbd91d986ec3703848d72aea5030e320d4988eb02aa3f9d48   Up 2 minutes   2 minutes ago   docker.io/library/redis:alpine   runc      "sh"
foo    18592900006405ee64788bd108ef1de3d24dc3add73725891f4787d0f8e036f5   Up 2 minutes   2 minutes ago   docker.io/library/redis:alpine   runc      "sh"
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
