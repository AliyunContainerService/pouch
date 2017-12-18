package main

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/alibaba/pouch/apis/types"
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
		display.AddRow([]string{c.Names[0], c.ID[:6], c.Status, c.Image, c.HostConfig.Runtime, createdTime(c.Created)})
	}
	display.Flush()
	return nil
}

// psExample shows examples in ps command, and is used in auto-generated cli docs.
func psExample() string {
	return `$ pouch ps
Name     ID       Status    Image                              Runtime   Created
1dad17   1dad17   stopped   docker.io/library/busybox:latest   runv      about 1 hour ago
505571   505571   stopped   docker.io/library/busybox:latest   runc      about 1 hour ago
`
}

// createdTime is used to show the time from creation to now.
func createdTime(created string) (s string) {
	start, _ := strconv.ParseInt(created, 10, 64)
	now := time.Now().Unix()
	diff := int(now - start)

	if diff >= 3600*24 {
		day := diff / (3600 * 24)
		s = "about " + strconv.Itoa(day) + " day"
		if day > 1 {
			s += "s"
		}
		s += " ago"
	} else if diff >= 3600 {
		hour := diff / 3600
		s = "about " + strconv.Itoa(hour) + " hour"
		if hour > 1 {
			s += "s"
		}
		s += " ago"
	} else if diff >= 60 {
		minute := diff / 60
		s = "about " + strconv.Itoa(minute) + " minute"
		if minute > 1 {
			s += "s"
		}
		s += " ago"
	} else {
		s = "about " + strconv.Itoa(diff) + " second"
		if diff > 1 {
			s += "s"
		}
		s += " ago"
	}

	return s
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
	ivalue, _ := strconv.ParseInt(c[i].Created, 10, 64)
	jvalue, _ := strconv.ParseInt(c[j].Created, 10, 64)
	return ivalue > jvalue
}
