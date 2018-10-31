package mgr

import (
	"fmt"

	"github.com/alibaba/pouch/apis/types"

	"github.com/docker/libnetwork"
)

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
