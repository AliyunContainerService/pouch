package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// portDescription is used to describe port command in detail and auto generate command doc.
var portDescription = "Return port binding information on Pouch container"

// PortCommand is used to implement 'port' command.
type PortCommand struct {
	baseCommand
	container string
	port      string
}

// Init initializes PortCommand command.
func (p *PortCommand) Init(c *Cli) {
	p.cli = c
	p.cmd = &cobra.Command{
		Use:   "port CONTAINER [PRIVATE_PORT[/PROTO]]",
		Short: "List port mappings or a specific mapping for the container",
		Long:  portDescription,
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			p.container = args[0]
			if len(args) > 1 {
				p.port = args[1]
			}
			return p.runPort()
		},
		Example: portExample(),
	}
}

// runPort is the entry of PortCommand command.
func (p *PortCommand) runPort() error {
	ctx := context.Background()
	apiClient := p.cli.Client()

	c, err := apiClient.ContainerGet(ctx, p.container)
	if err != nil {
		return err
	}

	if p.port != "" {
		port := p.port
		proto := "tcp"
		parts := strings.SplitN(port, "/", 2)

		if len(parts) == 2 && len(parts[1]) != 0 {
			port = parts[0]
			proto = parts[1]
		}
		natPort := port + "/" + proto
		newP, err := nat.NewPort(proto, port)
		if err != nil {
			return err
		}
		if portBindings, exists := c.NetworkSettings.Ports[string(newP)]; exists && portBindings != nil {
			for _, pb := range portBindings {
				fmt.Fprintf(os.Stdout, "%s:%s\n", pb.HostIP, pb.HostPort)
			}
			return nil
		}
		return errors.Errorf("No public port '%s' published for %s", natPort, p.container)
	}

	for from, portBindings := range c.NetworkSettings.Ports {
		for _, pb := range portBindings {
			fmt.Fprintf(os.Stdout, "%s -> %s:%s\n", from, pb.HostIP, pb.HostPort)
		}
	}

	return nil
}

// portExample shows examples in port command, and is used in auto-generated cli docs.
func portExample() string {
	return `$ pouch run -d -p 6379:6379 -p 6380:6380/udp  redis:latest
179eba2c29fb27a000bcda75cb2be271d1833ab140d1133799d0d4d865abc44e
$ pouch port 179
6379/tcp -> 0.0.0.0:6379
6380/udp -> 0.0.0.0:6380
$ pouch port 179 6379
0.0.0.0:6379
$ pouch port 179 6380/udp
0.0.0.0:6380
`
}
