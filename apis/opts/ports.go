package opts

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/go-connections/nat"
)

// ParseExposedPorts parse ports.
// FIXME(ziren): add more explanation
func ParseExposedPorts(portList, expose []string) (map[string]interface{}, error) {
	// translate ports
	tmpPorts, _, err := nat.ParsePortSpecs(portList)
	if err != nil {
		return nil, err
	}

	ports := map[string]interface{}{}
	for n, p := range tmpPorts {
		ports[string(n)] = p
	}
	for _, e := range expose {
		if strings.Contains(e, ":") {
			return nil, fmt.Errorf("invalid port format for --expose: %s", e)
		}

		//support two formats for expose, original format <portnum>/[<proto>] or <startport-endport>/[<proto>]
		proto, port := nat.SplitProtoPort(e)
		//parse the start and end port and create a sequence of ports to expose
		//if expose a port, the start and end port are the same
		start, end, err := nat.ParsePortRange(port)
		if err != nil {
			return nil, fmt.Errorf("invalid range format for --expose: %s, error: %s", e, err)
		}
		for i := start; i <= end; i++ {
			p, err := nat.NewPort(proto, strconv.FormatUint(i, 10))
			if err != nil {
				return nil, err
			}
			if _, exists := ports[string(p)]; !exists {
				ports[string(p)] = struct{}{}
			}
		}
	}

	return ports, nil
}

// ValidateExposedPorts verify the correction of exposed ports.
func ValidateExposedPorts(ports map[string]interface{}) error {
	// TODO

	return nil
}
