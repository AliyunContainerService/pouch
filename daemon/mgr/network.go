package mgr

import (
	"context"
	"fmt"
	"net"
	"path"
	"strings"

	apitypes "github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/network"
	"github.com/alibaba/pouch/network/types"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/meta"
	"github.com/alibaba/pouch/pkg/randomid"

	"github.com/docker/go-connections/nat"
	"github.com/docker/libnetwork"
	nwconfig "github.com/docker/libnetwork/config"
	"github.com/docker/libnetwork/netlabel"
	"github.com/docker/libnetwork/options"
	networktypes "github.com/docker/libnetwork/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// NetworkMgr defines interface to manage container network.
type NetworkMgr interface {
	// Create is used to create network.
	Create(ctx context.Context, create apitypes.NetworkCreateConfig) (*types.Network, error)

	// Get returns the information of network that specified name/id.
	Get(ctx context.Context, name string) (*types.Network, error)

	// List returns all networks on this host.
	List(ctx context.Context, labels map[string]string) ([]*types.Network, error)

	// NetworkRemove is used to delete an existing network.
	Remove(ctx context.Context, name string) error

	// EndpointCreate is used to create network endpoint.
	EndpointCreate(ctx context.Context, endpoint *types.Endpoint) (string, error)

	// EndpointInfo returns the information of endpoint that specified name/id.
	EndpointInfo(ctx context.Context, name string) (*types.Endpoint, error)

	// EndpointList returns all endpoints.
	EndpointList(ctx context.Context) ([]*types.Endpoint, error)

	// EndpointRemove is used to remove network endpoint.
	EndpointRemove(ctx context.Context, endpoint *types.Endpoint) error

	// Controller returns the network controller.
	Controller() libnetwork.NetworkController
}

// NetworkManager is the default implement of interface NetworkMgr.
type NetworkManager struct {
	store      *meta.Store
	controller libnetwork.NetworkController
	config     network.Config
}

// NewNetworkManager creates a brand new network manager.
func NewNetworkManager(cfg *config.Config, store *meta.Store, ctrMgr ContainerMgr) (*NetworkManager, error) {
	// Create a new controller instance
	if cfg.NetworkConfig.MetaPath == "" {
		cfg.NetworkConfig.MetaPath = path.Dir(store.BaseDir)
	}

	if cfg.NetworkConfig.ExecRoot == "" {
		cfg.NetworkConfig.ExecRoot = network.DefaultExecRoot
	}

	// get active sandboxes
	ctrs, err := ctrMgr.List(context.Background(),
		func(c *Container) bool {
			return (c.IsRunning() || c.IsPaused()) && !isContainer(c.HostConfig.NetworkMode)
		}, &ContainerListOption{All: true})
	if err != nil {
		logrus.Errorf("failed to new network manager, can not get container list")
		return nil, errors.Wrap(err, "failed to get container list")
	}
	cfg.NetworkConfig.ActiveSandboxes = make(map[string]interface{})
	for _, c := range ctrs {
		endpoint := BuildContainerEndpoint(c)
		sbOptions, err := buildSandboxOptions(cfg.NetworkConfig, endpoint)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build sandbox options")
		}
		cfg.NetworkConfig.ActiveSandboxes[c.NetworkSettings.SandboxID] = sbOptions
	}

	ctlOptions, err := controllerOptions(cfg.NetworkConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build network options")
	}

	controller, err := libnetwork.New(ctlOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create network controller")
	}

	return &NetworkManager{
		store:      store,
		controller: controller,
		config:     cfg.NetworkConfig,
	}, nil
}

// Create is used to create network.
func (nm *NetworkManager) Create(ctx context.Context, create apitypes.NetworkCreateConfig) (*types.Network, error) {
	name := create.Name
	driver := create.NetworkCreate.Driver
	id := randomid.Generate()

	nwOptions, err := networkOptions(create)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build network's options")
	}

	if net, err := nm.controller.NetworkByName(name); err == nil && net != nil {
		return nil, errors.Wrap(errtypes.ErrAlreadyExisted, fmt.Sprintf("network %s already exists", name))
	}

	net, err := nm.controller.NewNetwork(driver, name, id, nwOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create network")
	}

	network := types.Network{
		Name:    name,
		ID:      id,
		Type:    driver,
		Network: net,
	}

	return &network, nil
}

// Get returns the information of network for specified string that represent network name or ID.
// If network name is given, the network with same name is returned.
// If prefix of network ID is given, the network with same prefix is returned.
func (nm *NetworkManager) Get(ctx context.Context, idName string) (*types.Network, error) {
	n, err := nm.GetNetworkByName(idName)
	if err != nil && !isNoSuchNetworkError(err) {
		return nil, err
	}

	if n != nil {
		return n, nil
	}

	n, err = nm.GetNetworkByPartialID(idName)
	if err != nil {
		return nil, err
	}
	return n, err
}

// List returns all networks on this host.
func (nm *NetworkManager) List(ctx context.Context, labels map[string]string) ([]*types.Network, error) {
	nw := nm.controller.Networks()
	var net []*types.Network
	for _, n := range nw {
		nm := &types.Network{
			Name:    n.Name(),
			ID:      n.ID(),
			Type:    n.Type(),
			Network: n,
		}
		net = append(net, nm)
	}
	return net, nil
}

// Remove is used to delete an existing network.
func (nm *NetworkManager) Remove(ctx context.Context, name string) error {
	nw, err := nm.controller.NetworkByName(name)
	if err != nil {
		if err == libnetwork.ErrNoSuchNetwork(name) {
			return errors.Wrap(errtypes.ErrNotfound, err.Error())
		}
		return err
	}
	if nw == nil {
		return nil
	}

	return nw.Delete()
}

// GetNetworkByName returns the information of network that specified name.
func (nm *NetworkManager) GetNetworkByName(name string) (*types.Network, error) {
	n, err := nm.controller.NetworkByName(name)
	if err != nil {
		return nil, err
	}
	return &types.Network{
		Name:    n.Name(),
		ID:      n.ID(),
		Type:    n.Type(),
		Network: n,
	}, nil
}

// GetNetworkByPartialID returns the information of network that ID starts with the given prefix.
// If there are not matching networks, it fails with ErrNotfound.
// If there are multiple matching networks, it fails with ErrTooMany.
func (nm *NetworkManager) GetNetworkByPartialID(partialID string) (*types.Network, error) {
	network, err := nm.controller.NetworkByID(partialID)
	if err == nil {
		return &types.Network{
			Name:    network.Name(),
			ID:      network.ID(),
			Type:    network.Type(),
			Network: network,
		}, nil
	}
	if !isNoSuchNetworkError(err) {
		return nil, err
	}
	matchedNetworks := nm.GetNetworksByPartialID(partialID)
	if len(matchedNetworks) == 0 {
		return nil, errors.Wrap(errtypes.ErrNotfound, "network: "+partialID)
	}
	if len(matchedNetworks) > 1 {
		return nil, errors.Wrap(errtypes.ErrTooMany, "network: "+partialID)
	}
	return matchedNetworks[0], nil
}

// GetNetworksByPartialID returns a list of networks that ID starts with the given prefix.
func (nm *NetworkManager) GetNetworksByPartialID(partialID string) []*types.Network {
	var matchedNetworks []*types.Network

	walker := func(nw libnetwork.Network) bool {
		if strings.HasPrefix(nw.ID(), partialID) {
			matchedNetwork := &types.Network{
				Name:    nw.Name(),
				ID:      nw.ID(),
				Type:    nw.Type(),
				Network: nw,
			}
			matchedNetworks = append(matchedNetworks, matchedNetwork)
		}
		return false
	}
	nm.controller.WalkNetworks(walker)
	return matchedNetworks
}

// isNoSuchNetworkError looks up the error type and returns a bool if it is ErrNoSuchNetwork or not.
func isNoSuchNetworkError(err error) bool {
	_, ok := err.(libnetwork.ErrNoSuchNetwork)
	return ok
}

// EndpointCreate is used to create network endpoint.
func (nm *NetworkManager) EndpointCreate(ctx context.Context, endpoint *types.Endpoint) (string, error) {
	containerID := endpoint.Owner
	network := endpoint.Name
	networkConfig := endpoint.NetworkConfig
	endpointConfig := endpoint.EndpointConfig

	logrus.Debugf("create endpoint for container [%s] on network [%s]", containerID, network)
	if networkConfig == nil || endpointConfig == nil {
		return "", fmt.Errorf("networkConfig or endpointConfig can not be nil")
	}

	n, err := nm.controller.NetworkByName(network)
	if err != nil {
		if err == libnetwork.ErrNoSuchNetwork(network) {
			return "", errors.Wrap(errtypes.ErrNotfound, err.Error())
		}
		return "", err
	}

	// create endpoint
	epOptions, err := endpointOptions(n, endpoint)
	if err != nil {
		return "", err
	}

	endpointName := containerID[:8]
	ep, err := n.CreateEndpoint(endpointName, epOptions...)
	if err != nil {
		return "", err
	}

	defer func() {
		if err != nil {
			if err := ep.Delete(true); err != nil {
				logrus.Errorf("could not delete endpoint %s after failing to create endpoint(%v)", ep.Name(), err)
			}
		}
	}()

	// create sandbox
	sb := nm.getNetworkSandbox(containerID)
	if sb == nil {
		sandboxOptions, err := buildSandboxOptions(nm.config, endpoint)
		if err != nil {
			return "", fmt.Errorf("failed to build sandbox options(%v)", err)
		}

		sb, err = nm.controller.NewSandbox(containerID, sandboxOptions...)
		if err != nil {
			return "", fmt.Errorf("failed to create sandbox(%v)", err)
		}
	}
	networkConfig.SandboxID = sb.ID()
	networkConfig.SandboxKey = sb.Key()

	// endpoint joins into sandbox
	joinOptions, err := joinOptions(endpoint)
	if err != nil {
		return "", err
	}
	if err := ep.Join(sb, joinOptions...); err != nil {
		return "", fmt.Errorf("failed to join sandbox(%v)", err)
	}

	// update endpoint settings
	epInfo := ep.Info()
	if epInfo.Gateway() != nil {
		endpointConfig.Gateway = epInfo.Gateway().String()
	}
	if epInfo.GatewayIPv6().To16() != nil {
		endpointConfig.IPV6Gateway = epInfo.GatewayIPv6().String()
	}
	endpoint.ID = ep.ID()
	endpointConfig.EndpointID = ep.ID()
	endpointConfig.NetworkID = n.ID()

	iface := epInfo.Iface()
	if iface != nil {
		if iface.Address() != nil {
			mask, _ := iface.Address().Mask.Size()
			endpointConfig.IPPrefixLen = int64(mask)
			endpointConfig.IPAddress = iface.Address().IP.String()
		}

		if iface.MacAddress() != nil {
			endpointConfig.MacAddress = iface.MacAddress().String()
		}
	}

	return endpointName, nil
}

// EndpointInfo returns the information of endpoint that specified name/id.
func (nm *NetworkManager) EndpointInfo(ctx context.Context, name string) (*types.Endpoint, error) {
	// TODO
	return nil, nil
}

// EndpointList returns all endpoints.
func (nm *NetworkManager) EndpointList(ctx context.Context) ([]*types.Endpoint, error) {
	// TODO
	return nil, nil
}

// EndpointRemove is used to remove network endpoint.
func (nm *NetworkManager) EndpointRemove(ctx context.Context, endpoint *types.Endpoint) error {
	var (
		ep libnetwork.Endpoint
	)

	sid := endpoint.NetworkConfig.SandboxID
	epConfig := endpoint.EndpointConfig

	logrus.Debugf("remove endpoint(%s) on network(%s)", epConfig.EndpointID, endpoint.Name)

	if sid == "" {
		return nil
	}

	// find endpoint in network and delete it.
	sb, err := nm.controller.SandboxByID(sid)
	if err != nil {
		return errors.Wrapf(err, "failed to get sandbox by id(%s)", sid)
	}
	if sb == nil {
		return errors.Errorf("failed to get sandbox by id(%s)", sid)
	}

	eplist := sb.Endpoints()
	if len(eplist) == 0 {
		return errors.Errorf("no endpoint in sandbox(%s)", sid)
	}

	for _, e := range eplist {
		if e.ID() == epConfig.EndpointID {
			ep = e
			break
		}
	}

	if ep == nil {
		return errors.Errorf("not connected to the network(%s)", endpoint.Name)
	}

	if err := ep.Leave(sb); err != nil {
		return errors.Wrapf(err, "failed to leave network(%s)", endpoint.Name)
	}

	if err := ep.Delete(false); err != nil {
		return errors.Wrapf(err, "failed to delete endpoint(%s)", endpoint.ID)
	}

	// clean endpoint configure data
	nm.cleanEndpointConfig(epConfig)

	// check sandbox has endpoint or not.
	eplist = sb.Endpoints()
	if len(eplist) == 0 {
		if err := sb.Delete(); err != nil {
			logrus.Errorf("failed to delete sandbox id(%s), err(%v)", sid, err)
			return errors.Wrapf(err, "failed to delete sandbox id(%s)", sid)
		}
	}

	return nil
}

// Controller returns the network controller.
func (nm *NetworkManager) Controller() libnetwork.NetworkController {
	return nm.controller
}

func controllerOptions(cfg network.Config) ([]nwconfig.Option, error) {
	// TODO: parse network control config.
	options := []nwconfig.Option{}

	if cfg.MetaPath != "" {
		options = append(options, nwconfig.OptionDataDir(cfg.MetaPath))
	}

	if cfg.ExecRoot != "" {
		options = append(options, nwconfig.OptionExecRoot(cfg.ExecRoot))
	}

	if len(cfg.ActiveSandboxes) != 0 {
		options = append(options, nwconfig.OptionActiveSandboxes(cfg.ActiveSandboxes))
	}

	options = append(options, nwconfig.OptionDefaultDriver("bridge"))
	options = append(options, nwconfig.OptionDefaultNetwork("bridge"))

	// set bridge options
	options = append(options, bridgeDriverOptions(cfg.BridgeConfig))

	return options, nil
}

func bridgeDriverOptions(cfg network.BridgeConfig) nwconfig.Option {
	bridgeConfig := options.Generic{
		"EnableIPForwarding":  cfg.IPForward,
		"EnableIPTables":      cfg.IPTables,
		"EnableUserlandProxy": cfg.UserlandProxy}
	bridgeOption := options.Generic{netlabel.GenericData: bridgeConfig}

	return nwconfig.OptionDriverConfig("bridge", bridgeOption)
}

func networkOptions(create apitypes.NetworkCreateConfig) ([]libnetwork.NetworkOption, error) {
	// TODO: parse network config.
	networkCreate := create.NetworkCreate
	nwOptions := []libnetwork.NetworkOption{
		libnetwork.NetworkOptionEnableIPv6(networkCreate.EnableIPV6),
		libnetwork.NetworkOptionLabels(networkCreate.Labels),
	}

	// parse options
	if v, ok := networkCreate.Options["persist"]; ok && v == "true" {
		nwOptions = append(nwOptions, libnetwork.NetworkOptionPersist(true))
		delete(networkCreate.Options, "persist")
	}
	if v, ok := networkCreate.Options["dynamic"]; ok && v == "true" {
		nwOptions = append(nwOptions, libnetwork.NetworkOptionDynamic())
		delete(networkCreate.Options, "dynamic")
	}
	nwOptions = append(nwOptions, libnetwork.NetworkOptionDriverOpts(networkCreate.Options))

	if create.Name == "ingress" {
		nwOptions = append(nwOptions, libnetwork.NetworkOptionIngress(true))
	}

	if networkCreate.Internal {
		nwOptions = append(nwOptions, libnetwork.NetworkOptionInternalNetwork())
	}

	if networkCreate.IPAM != nil {
		ipam := networkCreate.IPAM
		v4Conf, v6Conf, err := getIpamConfig(ipam.Config)
		if err != nil {
			return nil, err
		}
		nwOptions = append(nwOptions, libnetwork.NetworkOptionIpam(ipam.Driver, "", v4Conf, v6Conf, ipam.Options))
	}

	return nwOptions, nil
}

func getIpamConfig(data []apitypes.IPAMConfig) ([]*libnetwork.IpamConf, []*libnetwork.IpamConf, error) {
	ipamV4Cfg := []*libnetwork.IpamConf{}
	ipamV6Cfg := []*libnetwork.IpamConf{}
	for _, d := range data {
		iCfg := libnetwork.IpamConf{}
		iCfg.PreferredPool = d.Subnet
		iCfg.SubPool = d.IPRange
		iCfg.Gateway = d.Gateway
		iCfg.AuxAddresses = d.AuxAddress
		ip, _, err := net.ParseCIDR(d.Subnet)
		if err != nil {
			return nil, nil, fmt.Errorf("Invalid subnet(%s), err(%v)", d.Subnet, err)
		}
		if ip.To4() != nil {
			ipamV4Cfg = append(ipamV4Cfg, &iCfg)
		} else {
			ipamV6Cfg = append(ipamV6Cfg, &iCfg)
		}
	}
	return ipamV4Cfg, ipamV6Cfg, nil
}

func (nm *NetworkManager) getNetworkSandbox(id string) libnetwork.Sandbox {
	var sb libnetwork.Sandbox
	nm.controller.WalkSandboxes(func(s libnetwork.Sandbox) bool {
		if s.ContainerID() == id {
			sb = s
			return true
		}
		return false
	})
	return sb
}

func endpointOptions(n libnetwork.Network, endpoint *types.Endpoint) ([]libnetwork.EndpointOption, error) {
	var createOptions []libnetwork.EndpointOption
	epConfig := endpoint.EndpointConfig
	if epConfig != nil {
		ipam := epConfig.IPAMConfig
		if ipam != nil && (ipam.IPV4Address != "" || ipam.IPV6Address != "" || len(ipam.LinkLocalIps) > 0) {
			var ipList []net.IP
			for _, ips := range ipam.LinkLocalIps {
				if ip := net.ParseIP(ips); ip != nil {
					ipList = append(ipList, ip)
				}
			}
			createOptions = append(createOptions,
				libnetwork.CreateOptionIpam(net.ParseIP(ipam.IPV4Address), net.ParseIP(ipam.IPV6Address), ipList, nil))
		}

		for _, alias := range epConfig.Aliases {
			createOptions = append(createOptions, libnetwork.CreateOptionMyAlias(alias))
		}
	}

	genericOption := options.Generic{}
	createOptions = append(createOptions, libnetwork.EndpointOptionGeneric(genericOption))

	if len(endpoint.GenericParams) > 0 {
		createOptions = append(createOptions, libnetwork.EndpointOptionGeneric(endpoint.GenericParams))
	}
	if endpoint.DisableResolver {
		createOptions = append(createOptions, libnetwork.CreateOptionDisableResolution())
	}

	return createOptions, nil
}

func buildSandboxOptions(config network.Config, endpoint *types.Endpoint) ([]libnetwork.SandboxOption, error) {
	var (
		sandboxOptions []libnetwork.SandboxOption
		dns            []string
		dnsSearch      []string
		dnsOptions     []string
	)

	sandboxOptions = append(sandboxOptions, libnetwork.OptionHostname(string(endpoint.Hostname)), libnetwork.OptionDomainname(endpoint.Domainname))

	if IsHost(endpoint.NetworkMode) {
		sandboxOptions = append(sandboxOptions, libnetwork.OptionUseDefaultSandbox())
		if len(endpoint.ExtraHosts) == 0 {
			sandboxOptions = append(sandboxOptions, libnetwork.OptionOriginHostsPath("/etc/hosts"))
		}
		if len(endpoint.DNS) == 0 && len(config.DNS) == 0 &&
			len(endpoint.DNSSearch) == 0 && len(config.DNSSearch) == 0 &&
			len(endpoint.DNSOptions) == 0 && len(config.DNSOptions) == 0 {
			sandboxOptions = append(sandboxOptions, libnetwork.OptionOriginResolvConfPath("/etc/resolv.conf"))
		}
	} else {
		sandboxOptions = append(sandboxOptions, libnetwork.OptionUseExternalKey())
	}

	sandboxOptions = append(sandboxOptions, libnetwork.OptionHostsPath(endpoint.HostsPath))
	sandboxOptions = append(sandboxOptions, libnetwork.OptionResolvConfPath(endpoint.ResolvConfPath))

	// parse DNS
	if len(endpoint.DNS) > 0 {
		dns = endpoint.DNS
	} else if len(config.DNS) > 0 {
		dns = config.DNS
	}
	for _, d := range dns {
		sandboxOptions = append(sandboxOptions, libnetwork.OptionDNS(d))
	}

	// parse DNS Search
	if len(endpoint.DNSSearch) > 0 {
		dnsSearch = endpoint.DNSSearch
	} else if len(config.DNSSearch) > 0 {
		dnsSearch = config.DNSSearch
	}
	for _, ds := range dnsSearch {
		sandboxOptions = append(sandboxOptions, libnetwork.OptionDNSSearch(ds))
	}

	// parse DNS Options
	if len(endpoint.DNSOptions) > 0 {
		dnsOptions = endpoint.DNSOptions
	} else if len(config.DNSOptions) > 0 {
		dnsOptions = config.DNSOptions
	}
	for _, ds := range dnsOptions {
		sandboxOptions = append(sandboxOptions, libnetwork.OptionDNSOptions(ds))
	}

	// TODO: secondary ip address
	// TODO: parse extra hosts
	// TODO: port mapping
	var bindings = make(nat.PortMap)
	if endpoint.PortBindings != nil {
		for p, b := range endpoint.PortBindings {
			bindings[nat.Port(p)] = []nat.PortBinding{}
			for _, bb := range b {
				bindings[nat.Port(p)] = append(bindings[nat.Port(p)], nat.PortBinding{
					HostIP:   bb.HostIP,
					HostPort: bb.HostPort,
				})
			}
		}
	}

	portSpecs := endpoint.ExposedPorts
	var ports = make([]nat.Port, len(portSpecs))
	var i int
	for p := range endpoint.ExposedPorts {
		ports[i] = nat.Port(p)
		i++
	}
	nat.SortPortMap(ports, bindings)

	var (
		exposeList []networktypes.TransportPort
		pbList     []networktypes.PortBinding
	)
	for _, port := range ports {
		expose := networktypes.TransportPort{}
		expose.Proto = networktypes.ParseProtocol(port.Proto())
		expose.Port = uint16(port.Int())
		exposeList = append(exposeList, expose)

		pb := networktypes.PortBinding{Port: expose.Port, Proto: expose.Proto}
		binding := bindings[port]
		for i := 0; i < len(binding); i++ {
			pbCopy := pb.GetCopy()
			newP, err := nat.NewPort(nat.SplitProtoPort(binding[i].HostPort))
			var portStart, portEnd int
			if err == nil {
				portStart, portEnd, err = newP.Range()
			}
			if err != nil {
				return nil, fmt.Errorf("failed to parsing HostPort value(%s):%v", binding[i].HostPort, err)
			}
			pbCopy.HostPort = uint16(portStart)
			pbCopy.HostPortEnd = uint16(portEnd)
			pbCopy.HostIP = net.ParseIP(binding[i].HostIP)
			pbList = append(pbList, pbCopy)
		}

		if endpoint.PublishAllPorts && len(binding) == 0 {
			pbList = append(pbList, pb)
		}
	}

	sandboxOptions = append(sandboxOptions,
		libnetwork.OptionPortMapping(pbList),
		libnetwork.OptionExposedPorts(exposeList))

	return sandboxOptions, nil
}

func (nm *NetworkManager) cleanEndpointConfig(epConfig *apitypes.EndpointSettings) {
	epConfig.EndpointID = ""
	epConfig.Gateway = ""
	epConfig.IPAddress = ""
	epConfig.IPPrefixLen = 0
	epConfig.IPV6Gateway = ""
	epConfig.GlobalIPV6Address = ""
	epConfig.GlobalIPV6PrefixLen = 0
	epConfig.MacAddress = ""
}

func joinOptions(endpoint *types.Endpoint) ([]libnetwork.EndpointOption, error) {
	var joinOptions []libnetwork.EndpointOption
	// TODO: parse endpoint's links

	// set priority option
	joinOptions = append(joinOptions, libnetwork.JoinOptionPriority(nil, endpoint.Priority))
	return joinOptions, nil
}
