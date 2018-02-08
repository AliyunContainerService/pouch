package bridge

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/extra/libnetwork/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/alibaba/pouch/network"
	"github.com/alibaba/pouch/pkg/errtypes"

	"github.com/docker/libnetwork/drivers/bridge"
	"github.com/docker/libnetwork/netlabel"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
)

// New is used to initialize bridge network.
func New(ctx context.Context, config network.BridgeConfig, manager mgr.NetworkMgr) error {
	// TODO: use config to set bridge network.
	// TODO: support ipv6.

	// clear exist bridge network
	if err := manager.Remove(ctx, "bridge"); err != nil {
		if !errtypes.IsNotfound(err) {
			return err
		}
	}

	// init host bridge network.
	if err := initBridgeDevice(); err != nil {
		return err
	}

	ipamV4Conf := &types.IPAMConfig{
		AuxAddress: make(map[string]string),
		Subnet:     DefaultSubnet,
		Gateway:    DefaultGateway,
	}

	ipam := &types.IPAM{
		Driver: "default",
		Config: []types.IPAMConfig{*ipamV4Conf},
	}

	networkCreate := types.NetworkCreate{
		Driver:     "bridge",
		EnableIPV6: false,
		Internal:   false,
		Options: map[string]string{
			bridge.BridgeName:         DefaultBridge,
			bridge.DefaultBridge:      strconv.FormatBool(true),
			netlabel.DriverMTU:        strconv.Itoa(1500),
			bridge.EnableICC:          strconv.FormatBool(true),
			bridge.DefaultBindingIP:   DefaultBindingIP,
			bridge.EnableIPMasquerade: strconv.FormatBool(true),
		},
		IPAM: ipam,
	}

	create := types.NetworkCreateConfig{
		Name:          "bridge",
		NetworkCreate: networkCreate,
	}

	_, err := manager.Create(ctx, create)
	return err
}

func initBridgeDevice() error {
	_, err := netlink.LinkByName(DefaultBridge)
	if err == nil {
		return nil
	}

	// generate mac address for bridge.
	var ip []int
	for _, v := range strings.Split(DefaultBridgeIP, ".") {
		tmp, _ := strconv.Atoi(v)
		ip = append(ip, tmp)
	}
	if len(ip) < 4 {
		return fmt.Errorf("bridge ip is invalid")
	}

	macAddr := fmt.Sprintf("02:42:%02x:%02x:%02x:%02x", ip[0], ip[1], ip[2], ip[3])
	logrus.Debugf("bridge mac address: %s", macAddr)

	la := netlink.NewLinkAttrs()
	la.HardwareAddr, err = net.ParseMAC(macAddr)
	if err != nil {
		return errors.Wrap(err, "failed to parse mac address")
	}

	la.Name = DefaultBridge

	b := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(b); err != nil {
		return errors.Wrap(err, "failed to add bridge device")
	}
	defer func() {
		if err != nil {
			netlink.LinkDel(b)
		}
	}()

	addr, err := netlink.ParseAddr(DefaultSubnet)
	if err != nil {
		return errors.Wrap(err, "failed to parse ip address")
	}
	if err := netlink.AddrAdd(b, addr); err != nil {
		return errors.Wrap(err, "failed to add ip address")
	}

	if err := netlink.LinkSetUp(b); err != nil {
		return errors.Wrap(err, "failed to set bridge device up")
	}

	return nil
}
