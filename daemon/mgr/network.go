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

	"github.com/docker/libnetwork"
	nwconfig "github.com/docker/libnetwork/config"
	"github.com/docker/libnetwork/netlabel"
	"github.com/docker/libnetwork/options"
	"github.com/pkg/errors"
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
	EndpointCreate(ctx context.Context, name string) error

	// EndpointRemove is used to create network endpoint.
	EndpointRemove(ctx context.Context, name string) error

	// EndpointList returns all endpoints.
	EndpointList(ctx context.Context) ([]*types.Endpoint, error)

	// EndpointInfo returns the information of endpoint that specified name/id.
	EndpointInfo(ctx context.Context, name string) (*types.Endpoint, error)
}

// NetworkManager is the default implement of interface NetworkMgr.
type NetworkManager struct {
	store      *meta.Store
	controller libnetwork.NetworkController
}

// NewNetworkManager creates a brand new network manager.
func NewNetworkManager(cfg *config.Config, store *meta.Store) (*NetworkManager, error) {
	// Create a new controller instance
	cfg.NetworkConfg.MetaPath = path.Dir(store.BaseDir)

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
func (nm *NetworkManager) EndpointCreate(ctx context.Context, name string) error {
	// TODO
	return nil
}

// EndpointRemove is used to create network endpoint.
func (nm *NetworkManager) EndpointRemove(ctx context.Context, name string) error {
	// TODO
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

func controllerOptions(cfg network.Config) ([]nwconfig.Option, error) {
	// TODO: parse network control config.
	options := []nwconfig.Option{}
	if cfg.MetaPath != "" {
		options = append(options, nwconfig.OptionDataDir(cfg.MetaPath))
	}

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
		libnetwork.NetworkOptionDriverOpts(networkCreate.Options),
		libnetwork.NetworkOptionLabels(networkCreate.Labels),
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
