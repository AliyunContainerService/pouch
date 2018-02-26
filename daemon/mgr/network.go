package mgr

import (
	"context"
	"fmt"
	"net"
	"path"

	apitypes "github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/meta"
	"github.com/alibaba/pouch/network"
	"github.com/alibaba/pouch/network/types"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/randomid"

	netlog "github.com/Sirupsen/logrus"
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

	// NetworkRemove is used to delete an existing network.
	Remove(ctx context.Context, name string) error

	// List returns all networks on this host.
	List(ctx context.Context, labels map[string]string) ([]*types.Network, error)

	// Get returns the information of network that specified name/id.
	Get(ctx context.Context, name string) (*types.Network, error)

	// EndpointCreate is used to create network endpoint.
	EndpointCreate(ctx context.Context, endpoint *types.Endpoint) (string, error)

	// EndpointRemove is used to remove network endpoint.
	EndpointRemove(ctx context.Context, endpoint *types.Endpoint) error

	// EndpointList returns all endpoints.
	EndpointList(ctx context.Context) ([]*types.Endpoint, error)

	// EndpointInfo returns the information of endpoint that specified name/id.
	EndpointInfo(ctx context.Context, name string) (*types.Endpoint, error)

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
func NewNetworkManager(cfg *config.Config, store *meta.Store) (*NetworkManager, error) {
	// Create a new controller instance
	cfg.NetworkConfg.MetaPath = path.Dir(store.BaseDir)
	cfg.NetworkConfg.ExecRoot = network.DefaultExecRoot

	initNetworkLog(cfg)

	ctlOptions, err := controllerOptions(cfg.NetworkConfg)
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
		config:     cfg.NetworkConfg,
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

// Get returns the information of network that specified name/id.
func (nm *NetworkManager) Get(ctx context.Context, name string) (*types.Network, error) {
	n, err := nm.controller.NetworkByName(name)
	if err != nil {
		if err == libnetwork.ErrNoSuchNetwork(name) {
			return nil, errors.Wrap(errtypes.ErrNotfound, err.Error())
		}
		return nil, err
	}

	if n != nil {
		return &types.Network{
			Name:    name,
			ID:      n.ID(),
			Type:    n.Type(),
			Network: n,
		}, nil
	}

	return nil, nil
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
	epOptions, err := endpointOptions(n, networkConfig, endpointConfig)
	if err != nil {
		return "", err
	}

	endpointName := containerID[:8]
	ep, err := n.CreateEndpoint(endpointName, epOptions...)
	if err != nil {
		logrus.Errorf("failed to create endpoint, err: %v", err)
		return "", err
	}

	// create sandbox
	sb := nm.getNetworkSandbox(containerID)
	if sb == nil {
		sandboxOptions, err := nm.sandboxOptions(endpoint)
		if err != nil {
			logrus.Errorf("failed to build sandbox options, err: %v", err)
			return "", err
		}

		sb, err = nm.controller.NewSandbox(containerID, sandboxOptions...)
		if err != nil {
			logrus.Errorf("failed to create sandbox, err: %v", err)
			return "", err
		}
	}
	networkConfig.SandboxID = sb.ID()
	networkConfig.SandboxKey = sb.Key()

	// endpoint joins into sandbox
	joinOptions, err := joinOptions(endpointConfig)
	if err != nil {
		return "", err
	}
	if err := ep.Join(sb, joinOptions...); err != nil {
		logrus.Errorf("failed to join sandbox, err: %v", err)
		return "", err
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
			endpointConfig.IPAddress = iface.Address().String()
		}

		if iface.MacAddress() != nil {
			endpointConfig.MacAddress = iface.MacAddress().String()
		}
	}

	return endpointName, nil
}

// EndpointRemove is used to remove network endpoint.
func (nm *NetworkManager) EndpointRemove(ctx context.Context, endpoint *types.Endpoint) error {
	sid := endpoint.NetworkConfig.SandboxID
	epConfig := endpoint.EndpointConfig

	logrus.Debugf("remove endpoint: %s on network: %s", epConfig.EndpointID, endpoint.Name)

	if sid == "" {
		return nil
	}

	sb, err := nm.controller.SandboxByID(sid)
	if err == nil {
		if err := sb.Delete(); err != nil {
			logrus.Errorf("failed to delete sandbox id: %s, err: %v", sid, err)
			return err
		}
	} else if _, ok := err.(networktypes.NotFoundError); !ok {
		logrus.Errorf("failed to get sandbox id: %s, err: %v", sid, err)
		return fmt.Errorf("failed to get sandbox id: %s, err: %v", sid, err)
	}

	// clean endpoint configure data
	nm.cleanEndpointConfig(epConfig)

	return nil
}

// EndpointList returns all endpoints.
func (nm *NetworkManager) EndpointList(ctx context.Context) ([]*types.Endpoint, error) {
	// TODO
	return nil, nil
}

// EndpointInfo returns the information of endpoint that specified name/id.
func (nm *NetworkManager) EndpointInfo(ctx context.Context, name string) (*types.Endpoint, error) {
	// TODO
	return nil, nil
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

	options = append(options, nwconfig.OptionDefaultDriver("bridge"))
	options = append(options, nwconfig.OptionDefaultNetwork("bridge"))

	// set bridge options
	options = append(options, bridgeDriverOptions())

	return options, nil
}

func bridgeDriverOptions() nwconfig.Option {
	bridgeConfig := options.Generic{
		"EnableIPForwarding":  true,
		"EnableIPTables":      true,
		"EnableUserlandProxy": true}
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
		nwOptions = append(nwOptions, libnetwork.NetworkOptionIngress())
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
			return nil, nil, fmt.Errorf("Invalid subnet %s : %v", d.Subnet, err)
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

func initNetworkLog(cfg *config.Config) {
	if cfg.Debug {
		netlog.SetLevel(netlog.DebugLevel)
	}

	formatter := &netlog.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000000000",
	}
	netlog.SetFormatter(formatter)
}

func endpointOptions(n libnetwork.Network, nwconfig *apitypes.NetworkSettings, epConfig *apitypes.EndpointSettings) ([]libnetwork.EndpointOption, error) {
	var createOptions []libnetwork.EndpointOption

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

	return createOptions, nil
}

func (nm *NetworkManager) sandboxOptions(endpoint *types.Endpoint) ([]libnetwork.SandboxOption, error) {
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
		if len(endpoint.DNS) == 0 && len(nm.config.DNS) == 0 &&
			len(endpoint.DNSSearch) == 0 && len(nm.config.DNSSearch) == 0 &&
			len(endpoint.DNSOptions) == 0 && len(nm.config.DNSOptions) == 0 {
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
	} else if len(nm.config.DNS) > 0 {
		dns = nm.config.DNS
	}
	for _, d := range dns {
		sandboxOptions = append(sandboxOptions, libnetwork.OptionDNS(d))
	}

	// parse DNS Search
	if len(endpoint.DNSSearch) > 0 {
		dnsSearch = endpoint.DNSSearch
	} else if len(nm.config.DNSSearch) > 0 {
		dnsSearch = nm.config.DNSSearch
	}
	for _, ds := range dnsSearch {
		sandboxOptions = append(sandboxOptions, libnetwork.OptionDNSSearch(ds))
	}

	// parse DNS Options
	if len(endpoint.DNSOptions) > 0 {
		dnsOptions = endpoint.DNSOptions
	} else if len(nm.config.DNSOptions) > 0 {
		dnsOptions = nm.config.DNSOptions
	}
	for _, ds := range dnsOptions {
		sandboxOptions = append(sandboxOptions, libnetwork.OptionDNSOptions(ds))
	}

	// TODO: secondary ip address
	// TODO: parse extra hosts
	// TODO: port mapping
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

func joinOptions(epConfig *apitypes.EndpointSettings) ([]libnetwork.EndpointOption, error) {
	var joinOptions []libnetwork.EndpointOption
	// TODO: parse endpoint's links

	return joinOptions, nil
}
