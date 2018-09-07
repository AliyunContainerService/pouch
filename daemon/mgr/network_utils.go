package mgr

import (
	"fmt"
	"strings"

	"github.com/alibaba/pouch/apis/types"

	"github.com/docker/libnetwork"
)

const (
	// NetworkModeHost means container network mode is host
	NetworkModeHost = "host"
)

// IsContainer is used to check if network mode is container mode.
func IsContainer(mode string) bool {
	parts := strings.SplitN(mode, ":", 2)
	return len(parts) > 1 && parts[0] == "container"
}

// IsHost is used to check if network mode is host mode.
func IsHost(mode string) bool {
	return mode == NetworkModeHost
}

// IsNone is used to check if network mode is none mode.
func IsNone(mode string) bool {
	return mode == "none"
}

// IsBridge is used to check if network mode is bridge mode.
func IsBridge(mode string) bool {
	return mode == "bridge"
}

// IsUserDefined is used to check if network mode is user-created.
func IsUserDefined(mode string) bool {
	return !IsBridge(mode) && !IsContainer(mode) && !IsHost(mode) && !IsNone(mode)
}

// IsDefault indicates whether container uses the default network stack.
func IsDefault(mode string) bool {
	return mode == "default"
}

// IsPrivate indicates whether container uses its private network stack.
func IsPrivate(mode string) bool {
	return !(IsHost(mode) || IsContainer(mode))
}

// hasUserDefinedIPAddress returns whether the passed endpoint configuration contains IP address configuration
func hasUserDefinedIPAddress(epConfig *types.EndpointSettings) bool {
	return epConfig != nil && epConfig.IPAMConfig != nil && (len(epConfig.IPAMConfig.IPV4Address) > 0 || len(epConfig.IPAMConfig.IPV6Address) > 0)
}

// User specified ip address is acceptable only for networks with user specified subnets.
func validateNetworkingConfig(network libnetwork.Network, epConfig *types.EndpointSettings) error {
	if network == nil || epConfig == nil {
		return nil
	}
	if !hasUserDefinedIPAddress(epConfig) {
		return nil
	}

	_, _, nwIPv4Configs, nwIPv6Configs := network.Info().IpamConfig()
	for _, s := range []struct {
		ipConfigured  bool
		subnetConfigs []*libnetwork.IpamConf
	}{
		{
			ipConfigured:  len(epConfig.IPAMConfig.IPV4Address) > 0,
			subnetConfigs: nwIPv4Configs,
		},
		{
			ipConfigured:  len(epConfig.IPAMConfig.IPV6Address) > 0,
			subnetConfigs: nwIPv6Configs,
		},
	} {
		if s.ipConfigured {
			foundSubnet := false
			for _, cfg := range s.subnetConfigs {
				if len(cfg.PreferredPool) > 0 {
					foundSubnet = true
					break
				}
			}
			if !foundSubnet {
				return fmt.Errorf("user specified IP address is supported only when connecting to networks with user configured subnets")
			}
		}
	}

	return nil
}
