package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/cli/inspect"

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
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("command 'pouch network %s' does not exist.\nPlease execute `pouch network --help` for more help", args[0])
		},
	}

	c.AddCommand(n, &NetworkCreateCommand{})
	c.AddCommand(n, &NetworkRemoveCommand{})
	c.AddCommand(n, &NetworkInspectCommand{})
	c.AddCommand(n, &NetworkListCommand{})
	c.AddCommand(n, &NetworkConnectCommand{})
	c.AddCommand(n, &NetworkDisconnectCommand{})
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
	format string
}

// Init initializes NetworkInspectCommand command.
func (n *NetworkInspectCommand) Init(c *Cli) {
	n.cli = c

	n.cmd = &cobra.Command{
		Use:   "inspect [OPTIONS] Network [Network...]",
		Short: "Inspect one or more pouch networks",
		Long:  networkInspectDescription,
		Args:  cobra.MinimumNArgs(1),
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
	n.cmd.Flags().StringVarP(&n.format, "format", "f", "", "Format the output using the given go template")
}

// runNetworkInspect is the entry of NetworkInspectCommand command.
func (n *NetworkInspectCommand) runNetworkInspect(args []string) error {
	ctx := context.Background()
	apiClient := n.cli.Client()

	getRefFunc := func(ref string) (interface{}, error) {
		return apiClient.NetworkInspect(ctx, ref)
	}

	return inspect.Inspect(os.Stdout, args, n.format, getRefFunc)
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
	respNetworkResource, err := apiClient.NetworkList(ctx)
	if err != nil {
		return err
	}

	display := n.cli.NewTableDisplay()
	display.AddRow([]string{"NETWORK ID", "NAME", "DRIVER", "SCOPE"})
	for _, network := range respNetworkResource {
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

// networkConnectDescription is used to describe network connect command in detail and auto generate command doc.
var networkConnectDescription = "Connect a container to a network in pouchd. " +
	"It must specify network's name and container's name."

// NetworkConnectCommand is used to implement 'network connect' command.
type NetworkConnectCommand struct {
	baseCommand

	ipAddress    string
	ipv6Address  string
	links        []string
	aliases      []string
	linklocalips []string
}

// Init initializes NetworkConnectCommand command.
func (n *NetworkConnectCommand) Init(c *Cli) {
	n.cli = c

	n.cmd = &cobra.Command{
		Use:   "connect [OPTIONS] NETWORK CONTAINER",
		Short: "Connect a container to a network",
		Long:  networkConnectDescription,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return n.runNetworkConnect(args)
		},
		Example: networkConnectExample(),
	}

	n.addFlags()
}

// addFlags adds flags for specific command.
func (n *NetworkConnectCommand) addFlags() {
	flagSet := n.cmd.Flags()

	flagSet.StringVar(&n.ipAddress, "ip", "", "IP Address")
	flagSet.StringVar(&n.ipv6Address, "ip6", "", "IPv6 Address")
	flagSet.StringSliceVar(&n.links, "link", []string{}, "Add link to another container")
	flagSet.StringSliceVar(&n.aliases, "alias", []string{}, "Add network-scoped alias for the container")
	flagSet.StringSliceVar(&n.linklocalips, "link-local-ip", []string{}, "Add a link-local address for the container")
}

// runNetworkConnect is the entry of NetworkConnectCommand command.
func (n *NetworkConnectCommand) runNetworkConnect(args []string) error {
	network := args[0]
	container := args[1]
	if network == "" {
		return fmt.Errorf("network name cannot be empty")
	}
	if container == "" {
		return fmt.Errorf("container name cannot be empty")
	}

	networkReq := &types.NetworkConnect{
		Container: container,
		EndpointConfig: &types.EndpointSettings{
			IPAMConfig: &types.EndpointIPAMConfig{
				IPV4Address:  n.ipAddress,
				IPV6Address:  n.ipv6Address,
				LinkLocalIps: n.linklocalips,
			},
			Links:   n.links,
			Aliases: n.aliases,
		},
	}

	ctx := context.Background()
	apiClient := n.cli.Client()
	err := apiClient.NetworkConnect(ctx, network, networkReq)
	if err != nil {
		return err
	}
	fmt.Printf("container %s is connected to network %s\n", container, network)

	return nil
}

// networkConnectExample shows examples in network connect command, and is used in auto-generated cli docs.
func networkConnectExample() string {
	return `$ pouch network connect net1 container1
container container1 is connected to network net1`
}

// networkDisconnectDescription is used to describe network disconnect command in detail and auto generate comand doc.
var networkDisconnectDescription = "Disconnect a container from a network"

// NetworkDisconnectCommand use to implement 'network disconnect' command, it disconnects given container from given network.
type NetworkDisconnectCommand struct {
	baseCommand
	force bool
}

// Init initializes 'network disconnect' command.
func (nd *NetworkDisconnectCommand) Init(c *Cli) {
	nd.cli = c
	nd.cmd = &cobra.Command{
		Use:   "disconnect [OPTIONS] NETWORK CONTAINER",
		Short: "Disconnect a container from a network",
		Args:  cobra.ExactArgs(2),
		Long:  networkDisconnectDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nd.runNetworkDisconnect(args)
		},
		Example: nd.networkDisconnectExample(),
	}
	nd.addFlags()
}

// addFlags adds flags for specific command.
func (nd *NetworkDisconnectCommand) addFlags() {
	// add flags
	nd.cmd.Flags().BoolVarP(&nd.force, "force", "f", false, "Force the container to disconnect from a network")
}

// runNetworkDisconnect is the entry of 'disconnect' command.
func (nd *NetworkDisconnectCommand) runNetworkDisconnect(args []string) error {
	network := args[0]
	container := args[1]

	if network == "" {
		return fmt.Errorf("network name cannot be empty")
	}
	if container == "" {
		return fmt.Errorf("container name cannot be empty")
	}

	ctx := context.Background()
	apiClient := nd.cli.Client()

	err := apiClient.NetworkDisconnect(ctx, network, container, nd.force)
	if err != nil {
		return err
	}
	fmt.Printf("container %s is disconnected from network %s\n successfully", container, network)

	return nil
}

// networkDisconnectExample shows examples in 'disconnect' command, and is used in auto-generated cli docs.
func (nd *NetworkDisconnectCommand) networkDisconnectExample() string {
	return `$ pouch network disconnect bridge test
container test is disconnected from network bridge successfully`
}
