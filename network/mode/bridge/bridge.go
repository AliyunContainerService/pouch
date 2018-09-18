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
	"github.com/alibaba/pouch/pkg/utils"

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

	// get bridge name
	bridgeName := DefaultBridge
	if config.Name != "" {
		bridgeName = config.Name
	}

	// get bridge ip
	bridgeIP := utils.StringDefault(config.IP, DefaultIPNet)
	ipNet, err := netlink.ParseIPNet(bridgeIP)
	if err != nil {
		return fmt.Errorf("failed to parse ip %v", bridgeIP)
	}
	logrus.Debugf("initialize bridge network, bridge ip: %s.", ipNet)

	// init host bridge network.
	_, err = initBridgeDevice(bridgeName, ipNet)
	if err != nil {
		return err
	}

	// get bridge subnet
	_, subnet, err := net.ParseCIDR(bridgeIP)
	if err != nil {
		return fmt.Errorf("failted to parse subnet %v", bridgeIP)
	}
	logrus.Debugf("initialize bridge network, bridge network: %s", subnet)

	// get ip range
	ipRange := DefaultIPRange
	if config.FixedCIDR != "" {
		ipRange = config.FixedCIDR
	} else {
		ipRange = subnet.String()
	}
	logrus.Debugf("initialize bridge network, bridge ip range in subnet: %s", ipRange)

	// get gateway
	gateway := DefaultGateway
	if config.GatewayIPv4 != "" {
		gateway = config.GatewayIPv4
	}
	logrus.Debugf("initialize bridge network, gateway: %s", gateway)

	ipamV4Conf := types.IPAMConfig{
		AuxAddress: make(map[string]string),
		Subnet:     subnet.String(),
		IPRange:    ipRange,
		Gateway:    gateway,
	}

	ipam := &types.IPAM{
		Driver: "default",
		Config: []types.IPAMConfig{ipamV4Conf},
	}

	mtu := network.DefaultNetworkMtu
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
			bridge.EnableIPMasquerade: strconv.FormatBool(true),
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

func containIP(ip net.IPNet, br netlink.Link) bool {
	addrs, err := netlink.AddrList(br, netlink.FAMILY_V4)
	if err == nil {
		for _, addr := range addrs {
			if ip.IP.Equal(addr.IP) {
				sizea, _ := ip.Mask.Size()
				sizeb, _ := addr.Mask.Size()
				if sizea == sizeb {
					return true
				}
			}
		}
	}
	return false
}

func existVethPair(br netlink.Link) bool {
	allLinks, err := netlink.LinkList()
	if err == nil {
		for _, l := range allLinks {
			if l.Type() == "veth" && l.Attrs().MasterIndex == br.Attrs().Index {
				return true
			}
		}
	}
	return false
}

func initBridgeDevice(name string, ipNet *net.IPNet) (netlink.Link, error) {
	br, err := netlink.LinkByName(name)
	if err == nil && br != nil {
		if containIP(*ipNet, br) { // do nothing if ip exists
			return br, nil
		}
		if existVethPair(br) {
			return nil, fmt.Errorf("failed to remove old bridge device due to existing veth pair")
		}
		netlink.LinkDel(br)
	}

	// generate mac address for bridge.
	var ip []int
	for _, v := range strings.Split(ipNet.IP.String(), ".") {
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

	addr, err := netlink.ParseAddr(ipNet.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse ip address")
	}
	if err := netlink.AddrAdd(b, addr); err != nil {
		return nil, errors.Wrap(err, "failed to add ip address")
	}

	if err := netlink.LinkSetUp(b); err != nil {
		return nil, errors.Wrap(err, "failed to set bridge device up")
	}

	br, err = netlink.LinkByName(name)
	if err != nil {
		return nil, err
	}

	return br, nil
}
