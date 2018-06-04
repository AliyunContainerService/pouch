package opts

import (
	"fmt"
	"net"
	"strings"

	"github.com/alibaba/pouch/apis/types"
)

// ParseNetworks parses network configurations of container.
func ParseNetworks(networks []string) (*types.NetworkingConfig, string, error) {
	var networkMode string
	if len(networks) == 0 {
		networkMode = "bridge"
	}
	networkingConfig := &types.NetworkingConfig{
		EndpointsConfig: map[string]*types.EndpointSettings{},
	}
	for _, network := range networks {
		name, parameter, mode, err := parseNetwork(network)
		if err != nil {
			return nil, "", err
		}

		if networkMode == "" || mode == "mode" {
			networkMode = name
		}

		if name == "container" {
			networkMode = fmt.Sprintf("%s:%s", name, parameter)
		} else if ipaddr := net.ParseIP(parameter); ipaddr != nil {
			networkingConfig.EndpointsConfig[name] = &types.EndpointSettings{
				IPAddress: parameter,
				IPAMConfig: &types.EndpointIPAMConfig{
					IPV4Address: parameter,
				},
			}
		}
	}

	return networkingConfig, networkMode, nil
}

// network format as below:
// [network]:[ip_address], such as: mynetwork:172.17.0.2 or mynetwork(ip alloc by ipam) or 172.17.0.2(default network is bridge)
// [network_mode]:[parameter], such as: host(use host network) or container:containerID(use exist container network)
// [network_mode]:[parameter]:mode, such as: mynetwork:172.17.0.2:mode(if the container has multi-networks, the network is the default network mode)
func parseNetwork(network string) (string, string, string, error) {
	var (
		name      string
		parameter string
		mode      string
	)
	if network == "" {
		return "", "", "", fmt.Errorf("invalid network: cannot be empty")
	}
	arr := strings.Split(network, ":")
	switch len(arr) {
	case 1:
		if ipaddr := net.ParseIP(arr[0]); ipaddr != nil {
			parameter = arr[0]
		} else {
			name = arr[0]
		}
	case 2:
		name = arr[0]
		if name == "container" {
			parameter = arr[1]
		} else if ipaddr := net.ParseIP(arr[1]); ipaddr != nil {
			parameter = arr[1]
		} else {
			mode = arr[1]
		}
	default:
		name = arr[0]
		parameter = arr[1]
		mode = arr[2]
	}

	return name, parameter, mode, nil
}

// ValidateNetworks verifies network configurations of container.
func ValidateNetworks(nwConfig *types.NetworkingConfig) error {
	if nwConfig == nil || len(nwConfig.EndpointsConfig) == 0 {
		return nil
	}

	// FIXME(ziren): this limitation may be removed in future
	// Now not create more then one interface for container is not allowed
	if len(nwConfig.EndpointsConfig) > 1 {
		l := make([]string, 0, len(nwConfig.EndpointsConfig))
		for k := range nwConfig.EndpointsConfig {
			l = append(l, k)
		}

		return fmt.Errorf("Container cannot be connected to network endpoints: %s", strings.Join(l, ", "))
	}

	// len(nwConfig.EndpointsConfig) == 1
	for k, v := range nwConfig.EndpointsConfig {
		if v == nil {
			return fmt.Errorf("no EndpointSettings for %s", k)
		}
		if v.IPAMConfig != nil {
			if v.IPAMConfig.IPV4Address != "" && net.ParseIP(v.IPAMConfig.IPV4Address).To4() == nil {
				return fmt.Errorf("invalid IPv4 address: %s", v.IPAMConfig.IPV4Address)
			}
			// TODO: check IPv6Address
		}
	}

	return nil
}
