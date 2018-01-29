package main

import (
	"fmt"

	"github.com/alibaba/pouch/apis/types"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// networkDescription defines the network command description and auto generate command doc.
var networkDescription = "Manager the networks in pouchd. " +
	"It contains the functions of create/remove/list/inspect network, 'driver' is used to list drivers that pouch support. " +
	"Now bridge network is supported in pouchd defaulted, it will be initialized when pouchd starting."

// NetworkCommand is used to implement 'network' command.
type NetworkCommand struct {
	baseCommand
}

// Init initializes NetworkCommand command.
func (n *NetworkCommand) Init(c *Cli) {
	n.cli = c

	n.cmd = &cobra.Command{
		Use:   "network [command]",
		Short: "Manage pouch networks",
		Long:  networkDescription,
		Args:  cobra.MinimumNArgs(1),
	}

	c.AddCommand(n, &NetworkCreateCommand{})
	c.AddCommand(n, &NetworkRemoveCommand{})
	c.AddCommand(n, &NetworkInspectCommand{})
	c.AddCommand(n, &NetworkListCommand{})
}

// networkCreateDescription is used to describe network create command in detail and auto generate command doc.
var networkCreateDescription = "Create a network in pouchd. " +
	"It must specify network's name and driver. You can use 'network driver' to get drivers that pouch support."

// NetworkCreateCommand is used to implement 'network create' command.
type NetworkCreateCommand struct {
	baseCommand

	name       string
	driver     string
	gateway    string
	ipRange    string
	ipamDriver string
	ipamOpts   []string
	subnet     []string
	options    []string
	labels     []string
}

// Init initializes NetworkCreateCommand command.
func (n *NetworkCreateCommand) Init(c *Cli) {
	n.cli = c

	n.cmd = &cobra.Command{
		Use:   "create [OPTIONS] [NAME]",
		Short: "Create a pouch network",
		Long:  networkCreateDescription,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return n.runNetworkCreate(args)
		},
		Example: networkCreateExample(),
	}

	n.addFlags()
}

// addFlags adds flags for specific command.
func (n *NetworkCreateCommand) addFlags() {
	flagSet := n.cmd.Flags()

	flagSet.StringVarP(&n.name, "name", "n", "", "the name of network")
	flagSet.StringVarP(&n.driver, "driver", "d", "bridge", "the driver of network")
	flagSet.StringVar(&n.gateway, "gateway", "", "the gateway of network")
	flagSet.StringVar(&n.ipRange, "ip-range", "", "the range of network's ip")
	flagSet.StringSliceVar(&n.subnet, "subnet", nil, "the subnet of network")
	flagSet.StringVar(&n.ipamDriver, "ipam-driver", "", "the ipam driver of network")
	flagSet.StringSliceVarP(&n.options, "option", "o", nil, "create network with options")
	flagSet.StringSliceVarP(&n.labels, "label", "l", nil, "create network with labels")
}

// runNetworkCreate is the entry of NetworkCreateCommand command.
func (n *NetworkCreateCommand) runNetworkCreate(args []string) error {
	name := n.name
	if len(args) != 0 {
		name = args[0]
	}
	if name == "" {
		return fmt.Errorf("network name cannot be empty")
	}

	networkRequest := &types.NetworkCreateConfig{
		Name: name,
	}

	if n.driver == "" {
		return fmt.Errorf("network driver cannot be empty")
	}
	networkCreate := types.NetworkCreate{
		Driver: n.driver,
	}

	networkRequest.NetworkCreate = networkCreate

	apiClient := n.cli.Client()
	resp, err := apiClient.NetworkCreate(networkRequest)
	if err != nil {
		return err
	}

	if resp.Warning != "" {
		fmt.Printf("WARNING: %s \n", resp.Warning)
	}
	fmt.Printf("%s: %s\n", name, resp.ID)

	return nil
}

// networkCreateExample shows examples in network create command, and is used in auto-generated cli docs.
func networkCreateExample() string {
	return `$ pouch network create -n pouchnet -d bridge --gateway 192.168.1.1 --ip-range 192.168.1.1/24 --subnet 192.168.1.1/24
pouchnet: e1d541722d68dc5d133cca9e7bd8fd9338603e1763096c8e853522b60d11f7b9`
}

// networkRemoveDescription is used to describe network remove command in detail and auto generate command doc.
var networkRemoveDescription = "Remove a network in pouchd. " +
	"It must specify network's name."

// NetworkRemoveCommand is used to implement 'network remove' command.
type NetworkRemoveCommand struct {
	baseCommand
}

// Init initializes NetworkRemoveCommand command.
func (n *NetworkRemoveCommand) Init(c *Cli) {
	n.cli = c

	n.cmd = &cobra.Command{
		Use:   "remove [OPTIONS] NAME",
		Short: "Remove a pouch network",
		Long:  networkRemoveDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return n.runNetworkRemove(args)
		},
		Example: networkRemoveExample(),
	}

	n.addFlags()
}

// addFlags adds flags for specific command.
func (n *NetworkRemoveCommand) addFlags() {
	//TODO add flags
}

// runNetworkRemove is the entry of NetworkRemoveCommand command.
func (n *NetworkRemoveCommand) runNetworkRemove(args []string) error {
	name := args[0]

	apiClient := n.cli.Client()
	if err := apiClient.NetworkRemove(name); err != nil {
		return err
	}
	fmt.Printf("Removed: %s\n", name)
	return nil
}

// networkRemoveExample shows examples in network remove command, and is used in auto-generated cli docs.
func networkRemoveExample() string {
	return `$ pouch network remove pouch-net
Removed: pouch-net`
}

// networkInspectDescription is used to describe network inspect command in detail and auto generate command doc.
var networkInspectDescription = "Inspect a network in pouchd. " +
	"It must specify network's name."

// NetworkInspectCommand is used to implement 'network inspect' command.
type NetworkInspectCommand struct {
	baseCommand
}

// Init initializes NetworkInspectCommand command.
func (n *NetworkInspectCommand) Init(c *Cli) {
	n.cli = c

	n.cmd = &cobra.Command{
		Use:   "inspect [OPTIONS] NAME",
		Short: "Inspect a pouch network",
		Long:  networkInspectDescription,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return n.runNetworkInspect(args)
		},
		Example: networkInspectExample(),
	}

	n.addFlags()
}

// addFlags adds flags for specific command.
func (n *NetworkInspectCommand) addFlags() {
	//TODO add flags
}

// runNetworkInspect is the entry of NetworkInspectCommand command.
func (n *NetworkInspectCommand) runNetworkInspect(args []string) error {
	name := args[0]

	apiClient := n.cli.Client()
	resp, err := apiClient.NetworkInspect(name)
	if err != nil {
		return err
	}

	n.cli.Print(resp)
	return nil
}

// networkInspectExample shows examples in network inspect command, and is used in auto-generated cli docs.
func networkInspectExample() string {
	return `$ pouch network inspect net1
Name:         net1
Scope:        
Driver:       bridge
EnableIPV6:   false
ID:           c33c2646dc8ce9162faa65d17e80582475bbe53dc70ba0dc4def4b71e44551d6
Internal:     false`
}

// networkListDescription is used to describe network list command in detail and auto generate command doc.
var networkListDescription = "List networks in pouchd. " +
	"It lists the network's Id, name, driver and scope."

// NetworkListCommand is used to implement 'network list' command.
type NetworkListCommand struct {
	baseCommand
}

// Init initializes NetworkListCommand command.
func (n *NetworkListCommand) Init(c *Cli) {
	n.cli = c

	n.cmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List pouch networks",
		Long:    networkListDescription,
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return n.runNetworkList(args)
		},
		Example: networkListExample(),
	}

	n.addFlags()
}

// addFlags adds flags for specific command.
func (n *NetworkListCommand) addFlags() {
	//TODO add flags
}

// runNetworkList is the entry of NetworkListCommand command.
func (n *NetworkListCommand) runNetworkList(args []string) error {
	logrus.Debugf("list the networks")

	apiClient := n.cli.Client()
	resp, err := apiClient.NetworkList()
	if err != nil {
		return err
	}

	display := n.cli.NewTableDisplay()
	display.AddRow([]string{"NETWORK ID", "NAME", "DRIVER", "SCOPE"})
	for _, network := range resp.Networks {
		display.AddRow([]string{
			network.ID[:10],
			network.Name,
			network.Driver,
			network.Scope,
		})
	}

	display.Flush()
	return nil
}

// networkListExample shows examples in network list command, and is used in auto-generated cli docs.
func networkListExample() string {
	return `$ pouch network ls
NETWORK ID   NAME   DRIVER    SCOPE
6f7aba8a58   net2   bridge
55f134176c   net3   bridge
e495f50913   net1   bridge
`
}
