package main

import (
	"context"
	"fmt"
	"strings"

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
	subnet     string
	enableIPv6 bool
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
	flagSet.StringVar(&n.subnet, "subnet", "", "the subnet of network")
	flagSet.StringVar(&n.ipamDriver, "ipam-driver", "default", "the ipam driver of network")
	flagSet.BoolVar(&n.enableIPv6, "enable-ipv6", false, "enable ipv6 network")
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

	networkRequest, err := n.buildNetworkCreateRequest(name)
	if err != nil {
		return err
	}

	ctx := context.Background()
	apiClient := n.cli.Client()
	resp, err := apiClient.NetworkCreate(ctx, networkRequest)
	if err != nil {
		return err
	}

	if resp.Warning != "" {
		fmt.Printf("WARNING: %s \n", resp.Warning)
	}
	fmt.Printf("%s: %s\n", name, resp.ID)

	return nil
}

func (n *NetworkCreateCommand) buildNetworkCreateRequest(name string) (*types.NetworkCreateConfig, error) {
	if n.driver == "" {
		return nil, fmt.Errorf("network driver cannot be empty")
	}

	options, err := parseSliceToMap(n.options)
	if err != nil {
		return nil, err
	}

	labels, err := parseSliceToMap(n.labels)
	if err != nil {
		return nil, err
	}

	ipamOptions, err := parseSliceToMap(n.ipamOpts)
	if err != nil {
		return nil, err
	}

	ipam := &types.IPAM{
		Driver:  n.ipamDriver,
		Options: ipamOptions,
		Config:  []types.IPAMConfig{},
	}

	if n.subnet != "" || n.gateway != "" || n.ipRange != "" {
		ipamConfig := types.IPAMConfig{
			AuxAddress: make(map[string]string),
			Subnet:     n.subnet,
			Gateway:    n.gateway,
			IPRange:    n.ipRange,
		}
		ipam.Config = append(ipam.Config, ipamConfig)
	}

	networkCreate := types.NetworkCreate{
		Driver:         n.driver,
		EnableIPV6:     n.enableIPv6,
		Internal:       false,
		CheckDuplicate: true,
		Options:        options,
		Labels:         labels,
		IPAM:           ipam,
	}
	networkRequest := &types.NetworkCreateConfig{
		Name:          name,
		NetworkCreate: networkCreate,
	}

	return networkRequest, nil
}

func parseSliceToMap(slices []string) (map[string]string, error) {
	maps := map[string]string{}

	for _, s := range slices {
		kv := strings.Split(s, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid slice format: %s", s)
		}
		maps[kv[0]] = kv[1]
	}
	return maps, nil
}

// networkCreateExample shows examples in network create command, and is used in auto-generated cli docs.
func networkCreateExample() string {
	return `$ pouch network create -n pouchnet -d bridge --gateway 192.168.1.1 --subnet 192.168.1.0/24
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
		Use:     "remove [OPTIONS] NAME",
		Aliases: []string{"rm"},
		Short:   "Remove a pouch network",
		Long:    networkRemoveDescription,
		Args:    cobra.ExactArgs(1),
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

	ctx := context.Background()
	apiClient := n.cli.Client()
	if err := apiClient.NetworkRemove(ctx, name); err != nil {
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

	ctx := context.Background()
	apiClient := n.cli.Client()
	resp, err := apiClient.NetworkInspect(ctx, name)
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

	ctx := context.Background()
	apiClient := n.cli.Client()
	resp, err := apiClient.NetworkList(ctx)
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
