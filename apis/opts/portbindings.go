package opts

import (
	"fmt"

	"github.com/alibaba/pouch/apis/types"

	"github.com/docker/go-connections/nat"
)

// ParsePortBinding parse string list to PortMap
// FIXME(ziren): add examples
func ParsePortBinding(ports []string) (types.PortMap, error) {
	// parse port binding
	_, tmpPortBindings, err := nat.ParsePortSpecs(ports)
	if err != nil {
		return nil, err
	}

	portBindings := make(types.PortMap)
	for n, pbs := range tmpPortBindings {
		portBindings[string(n)] = []types.PortBinding{}
		for _, tmpPb := range pbs {
			pb := types.PortBinding{HostIP: tmpPb.HostIP, HostPort: tmpPb.HostPort}
			portBindings[string(n)] = append(portBindings[string(n)], pb)
		}
	}

	return portBindings, nil
}

// ValidatePortBinding verify PortMap struct correctness.
func ValidatePortBinding(portBindings types.PortMap) error {
	for port := range portBindings {
		_, portStr := nat.SplitProtoPort(string(port))
		if _, err := nat.ParsePort(portStr); err != nil {
			return fmt.Errorf("invalid port specification: %q, err: %v", portStr, err)
		}

		for _, pb := range portBindings[port] {
			_, err := nat.NewPort(nat.SplitProtoPort(pb.HostPort))
			if err != nil {
				return fmt.Errorf("invalud port specification: %q, err: %v", pb.HostPort, err)
			}
		}
	}

	return nil
}
