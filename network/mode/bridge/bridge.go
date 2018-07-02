package bridge

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/network"
	"github.com/alibaba/pouch/pkg/errtypes"

	"github.com/docker/libnetwork/drivers/bridge"
	"github.com/docker/libnetwork/netlabel"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// New is used to initialize bridge network.
func New(ctx context.Context, config network.BridgeConfig, manager mgr.NetworkMgr) error {
	// TODO: support ipv6.

	// clear exist bridge network
	if err := manager.Remove(ctx, "bridge"); err != nil {
		if !errtypes.IsNotfound(err) {
			return err
		}
	}

	// set bridge name
	bridgeName := DefaultBridge
	if config.Name != "" {
		bridgeName = config.Name
	}

	// init host bridge network.
	br, err := initBridgeDevice(bridgeName)
	if err != nil {
		return err
	}

	var (
		bridgeIPv4Address string
	)
	Addrs, err := netlink.AddrList(br, netlink.FAMILY_V4)
	if err != nil {
		return errors.Wrap(err, "failed to get bridge addr")
	}
	for _, addr := range Addrs {
		cidr := addr.String()
		if strings.Contains(cidr, ":") {
			continue
		}

		parts := strings.Split(cidr, " ")
		if len(parts) != 2 {
			continue
		}

		bridgeIPv4Address = parts[0]
		break
	}

	// get subnet
	subnet := DefaultSubnet
	if config.IP != "" {
		subnet = config.IP
	} else if bridgeIPv4Address != "" {
		subnet = bridgeIPv4Address
	}
	logrus.Debugf("initialize bridge network, subnet: %s", subnet)

	// get ip range
	ipRange := subnet
	if config.FixedCIDR != "" {
		ipRange = config.FixedCIDR
	}
	logrus.Debugf("initialize bridge network, bridge ip range in subnet: %s", ipRange)

	// get gateway
	gateway := DefaultGateway
	if config.GatewayIPv4 != "" {
		gateway = config.GatewayIPv4
	} else {
		// get the default route set as gateway.
		routes, err := netlink.RouteList(br, netlink.FAMILY_V4)
		if err != nil {
			return errors.Wrap(err, "failed to get route list")
		}
		for _, route := range routes {
			gw := route.Gw.String()
			if gw != "" && gw != "<nil>" {
				gateway = gw
				break
			}
		}

		// nat mode bridge have no default route, so let the bridge ip as gateway.
		if bridgeIPv4Address != "" {
			gateway = strings.Split(bridgeIPv4Address, "/")[0]
		}
	}
	logrus.Debugf("initialize bridge network, gateway: %s", gateway)

	ipamV4Conf := types.IPAMConfig{
		AuxAddress: make(map[string]string),
		Subnet:     subnet,
		IPRange:    ipRange,
		Gateway:    gateway,
	}

	ipam := &types.IPAM{
		Driver: "default",
		Config: []types.IPAMConfig{ipamV4Conf},
	}

	mtu := 1500
	if config.Mtu != 0 {
		mtu = config.Mtu
	}

	networkCreate := types.NetworkCreate{
		Driver:     "bridge",
		EnableIPV6: false,
		Internal:   false,
		Options: map[string]string{
			bridge.BridgeName:         bridgeName,
			bridge.DefaultBridge:      strconv.FormatBool(true),
			netlabel.DriverMTU:        strconv.Itoa(mtu),
			bridge.EnableICC:          strconv.FormatBool(true),
			bridge.DefaultBindingIP:   DefaultBindingIP,
			bridge.EnableIPMasquerade: strconv.FormatBool(false),
		},
		IPAM: ipam,
	}

	create := types.NetworkCreateConfig{
		Name:          "bridge",
		NetworkCreate: networkCreate,
	}

	_, err = manager.Create(ctx, create)
	return err
}

func initBridgeDevice(name string) (netlink.Link, error) {
	br, err := netlink.LinkByName(name)
	if err == nil && br != nil {
		return br, nil
	}

	// generate mac address for bridge.
	var ip []int
	for _, v := range strings.Split(DefaultBridgeIP, ".") {
		tmp, _ := strconv.Atoi(v)
		ip = append(ip, tmp)
	}
	if len(ip) < 4 {
		return nil, fmt.Errorf("bridge ip is invalid")
	}

	macAddr := fmt.Sprintf("02:42:%02x:%02x:%02x:%02x", ip[0], ip[1], ip[2], ip[3])

	la := netlink.NewLinkAttrs()
	la.HardwareAddr, err = net.ParseMAC(macAddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse mac address")
	}

	la.Name = name

	b := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(b); err != nil {
		return nil, errors.Wrap(err, "failed to add bridge device")
	}
	defer func() {
		if err != nil {
			netlink.LinkDel(b)
		}
	}()

	addr, err := netlink.ParseAddr(DefaultSubnet)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse ip address")
	}
	if err := netlink.AddrAdd(b, addr); err != nil {
		return nil, errors.Wrap(err, "failed to add ip address")
	}

	if err := netlink.LinkSetUp(b); err != nil {
		return nil, errors.Wrap(err, "failed to set bridge device up")
	}

	br, err = netlink.LinkByName(DefaultBridge)
	if err != nil {
		return nil, err
	}

	return br, nil
}
