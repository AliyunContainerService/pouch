package bridge

import (
	"context"
	"strconv"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/network"

	"github.com/docker/libnetwork/drivers/bridge"
	"github.com/docker/libnetwork/netlabel"
	"github.com/vishvananda/netlink"
)

// New is used to initialize bridge network.
func New(ctx context.Context, config network.BridgeConfig, manager mgr.NetworkMgr) error {
	// TODO: use config to set bridge network.
	// TODO: support ipv6.

	if _, err := netlink.LinkByName(DefaultBridge); err != nil {
		la := netlink.NewLinkAttrs()
		la.Name = DefaultBridge
		b := &netlink.Bridge{LinkAttrs: la}
		netlink.LinkAdd(b)
	}

	// clear exist bridge network
	if err := manager.NetworkRemove(ctx, "bridge"); err != nil {
		return err
	}

	ipamV4Conf := &types.IPAMConfig{
		AuxAddress: map[string]string{},
		Subnet:     DefaultSubnet,
		IPRange:    DefaultIPRange,
		Gateway:    DefaultGateway,
	}

	ipam := &types.IPAM{
		Driver: "default",
		Config: []*types.IPAMConfig{ipamV4Conf},
	}

	networkCreate := &types.NetworkCreate{
		Driver:     "bridge",
		EnableIPV6: false,
		Options: map[string]string{
			bridge.BridgeName:       DefaultBridge,
			bridge.DefaultBridge:    strconv.FormatBool(true),
			netlabel.DriverMTU:      strconv.Itoa(0),
			bridge.EnableICC:        strconv.FormatBool(true),
			bridge.DefaultBindingIP: DefaultBridgeIP,
		},
		IPAM: ipam,
	}

	create := types.NetworkCreateConfig{
		Name:          "bridge",
		NetworkCreate: networkCreate,
	}

	_, err := manager.NetworkCreate(ctx, create)
	return err
}
