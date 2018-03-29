package main

import (
	"context"
	"os"

	"github.com/alibaba/pouch/cli/inspect"

	"github.com/spf13/cobra"
)

// inspectDescription is used to describe inspect command in detail and auto generate command doc.
var inspectDescription = "Return detailed information on Pouch container"

// InspectCommand is used to implement 'inspect' command.
type InspectCommand struct {
	baseCommand
	format string
}

// Init initializes InspectCommand command.
func (p *InspectCommand) Init(c *Cli) {
	p.cli = c
	p.cmd = &cobra.Command{
		Use:   "inspect [OPTIONS] CONTAINER",
		Short: "Get the detailed information of container",
		Long:  inspectDescription,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return inspect.MultiInspect(args, p.runInspect)
		},
		Example: inspectExample(),
	}
	p.addFlags()
}

// addFlags adds flags for specific command.
func (p *InspectCommand) addFlags() {
	p.cmd.Flags().StringVarP(&p.format, "format", "f", "", "Format the output using the given go template")
}

// runInspect is the entry of InspectCommand command.
func (p *InspectCommand) runInspect(args []string) error {
	ctx := context.Background()
	apiClient := p.cli.Client()
	name := args[0]

	getRefFunc := func(ref string) (interface{}, error) {
		return apiClient.ContainerGet(ctx, ref)
	}

	return inspect.Inspect(os.Stdout, name, p.format, getRefFunc)
}

// inspectExample shows examples in inspect command, and is used in auto-generated cli docs.
func inspectExample() string {
	return `$ pouch inspect 08e
{
  "Id": "08ee444faa3c6634ecdecea26de46e8a6a16efefd9afb72eb3457320b333fc60",
  "Created": "2017-12-04 14:48:59",
  "Path": "",
  "Args": null,
  "State": {
    "StartedAt": "0001-01-01T00:00:00Z",
    "Status": 0,
    "FinishedAt": "0001-01-01T00:00:00Z",
    "Pid": 25006,
    "ExitCode": 0,
    "Error": ""
  },
  "Image": "registry.docker-cn.com/library/centos:latest",
  "ResolvConfPath": "",
  "HostnamePath": "",
  "HostsPath": "",
  "LogPath": "",
  "Name": "08ee44",
  "RestartCount": 0,
  "Driver": "",
  "MountLabel": "",
  "ProcessLabel": "",
  "AppArmorProfile": "",
  "ExecIDs": null,
  "HostConfig": null,
  "HostRootPath": ""
}`
}
