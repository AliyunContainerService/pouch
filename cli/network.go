package main

import (
	"fmt"

	"github.com/alibaba/pouch/apis/types"

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

	name string
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
