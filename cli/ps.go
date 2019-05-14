package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"text/template"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/cli/formatter"
	"github.com/alibaba/pouch/pkg/utils/filters"

	"github.com/spf13/cobra"
)

// psDescription is used to describe ps command in detail and auto generate command doc.
var psDescription = "\nList Containers with container name, ID, status, creation time, image reference and runtime."
var psDefaultFormat = "table {{.Names}}\t{{.ID}}\t{{.Status}}\t{{.RunningFor}}\t{{.Image}}\t{{.Runtime}}\n"

// containerList is used to save the container list.
type containerList []*types.Container

// PsCommand is used to implement 'ps' command.
type PsCommand struct {
	baseCommand

	// flags for ps command
	flagAll     bool
	flagQuiet   bool
	flagNoTrunc bool
	flagFilter  []string
	flagFormat  string
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
	flagSet.StringVarP(&p.flagFormat, "format", "", "", "intelligent-print containers based on Go template")
	flagSet.StringSliceVarP(&p.flagFilter, "filter", "f", nil, "Filter output based on given conditions, support filter key [ id label name status ]")
}

// runPs is the entry of PsCommand command.
func (p *PsCommand) runPs(args []string) error {
	ctx := context.Background()
	apiClient := p.cli.Client()

	filter, err := filters.Parse(p.flagFilter)
	if err != nil {
		return err
	}

	var containers containerList

	option := types.ContainerListOptions{
		All:    p.flagAll,
		Filter: filter,
	}
	containers, err = apiClient.ContainerList(ctx, option)
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
	// add to format the output with go template
	format := p.flagFormat
	tmplH := template.New("ps_head")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, p.cli.padding, ' ', 0)
	if len(format) == 0 {
		format = psDefaultFormat
	}
	// true is table,false is raw
	tableOrRaw := formatter.IsTable(format)
	format = formatter.PreFormat(format)
	if tableOrRaw {
		containerHeader := formatter.ContainerHeader
		p.cli.FormatDisplay(format, tmplH, containerHeader, w)
	}
	for _, c := range containers {
		containerContext, err := formatter.NewContainerContext(c, p.flagNoTrunc)
		if err != nil {
			return err
		}
		tmplD := template.New("ps_detail")
		p.cli.FormatDisplay(format, tmplD, containerContext, w)
	}
	w.Flush()
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

$ pouch ps --no-trunc
Name   ID                                                                 Status        Created        Image                            Runtime
foo2   692c77587b38f60bbd91d986ec3703848d72aea5030e320d4988eb02aa3f9d48   Up 1 minute   1 minute ago   docker.io/library/redis:alpine   runc
foo    18592900006405ee64788bd108ef1de3d24dc3add73725891f4787d0f8e036f5   Up 1 minute   1 minute ago   docker.io/library/redis:alpine   runc

$ pouch ps --no-trunc -q
692c77587b38f60bbd91d986ec3703848d72aea5030e320d4988eb02aa3f9d48
18592900006405ee64788bd108ef1de3d24dc3add73725891f4787d0f8e036f5

$ pouch ps --no-trunc -a
Name   ID                                                                 Status         Created         Image                            Runtime
foo3   63fd6371f3d614bb1ecad2780972d5975ca1ab534ec280c5f7d8f4c7b2e9989d   created        2 minutes ago   docker.io/library/redis:alpine   runc
foo2   692c77587b38f60bbd91d986ec3703848d72aea5030e320d4988eb02aa3f9d48   Up 2 minutes   2 minutes ago   docker.io/library/redis:alpine   runc
foo    18592900006405ee64788bd108ef1de3d24dc3add73725891f4787d0f8e036f5   Up 2 minutes   2 minutes ago   docker.io/library/redis:alpine   runc

$ pouch ps --format "table {{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Command}}\t{{.CreatedAt}}\t{{.RunningFor}}\t{{.Ports}}\t{{.Status}}\t{{.Size}}\t{{.Labels}}\t{{.Mounts}}\t{{.LocalVolumes}}\t{{.Networks}}\t{{.Runtime}}\t{{.ImageID}}"
ID       Name     Image                                          Command   CreatedAt                                Created         Ports              Status         Size   Labels   Mounts        Volumes   Networks   Runtime   ImageID
869433   test   registry.hub.docker.com/library/busybox:1.28   sh        2019-05-29 05:40:46.64617376 +0000 UTC   6 seconds ago   3333/tcp->:3333;   Up 6 seconds   0B     a = b;   /root/test;   0         bridge     runc      sha256:8c811b4aec35f259572d0f79207bc0678df4c736eeec50bc9fec37ed936a472a
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
