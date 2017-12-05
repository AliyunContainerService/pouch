package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// inspectDescription is used to describe inspect command in detail and auto generate command doc.
var inspectDescription = "Return detailed information on Pouch container"

// InspectCommand is used to implement 'inspect' command.
type InspectCommand struct {
	baseCommand
}

// Init initializes InspectCommand command.
func (p *InspectCommand) Init(c *Cli) {
	p.cli = c
	p.cmd = &cobra.Command{
		Use:   "inspect [container]",
		Short: "Get the detailed information of container",
		Long:  inspectDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.runInpsect(args)
		},
		Example: inspectExample(),
	}
}

// runInpsect is the entry of InspectCommand command.
func (p *InspectCommand) runInpsect(args []string) error {
	apiClient := p.cli.Client()
	name := args[0]
	container, err := apiClient.ContainerGet(name)
	if err != nil {
		return err
	}

	containerjson, err := json.MarshalIndent(&container, "", "  ")
	if err != nil {
		return err
	}
	fmt.Print(string(containerjson) + "\n")
	return nil
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
