package types

import (
	"github.com/alibaba/pouch/apis/types"

	"github.com/go-openapi/strfmt"
)

// Endpoint defines the network endpoint struct.
type Endpoint struct {
	Name  string
	ID    string
	Owner string

	Hostname       strfmt.Hostname
	Domainname     string
	HostnamePath   string
	HostsPath      string
	ExtraHosts     []string
	ResolvConfPath string
	DNS            []string
	DNSOptions     []string
	DNSSearch      []string

	NetworkDisabled bool
	NetworkMode     string
	MacAddress      string
	ExposedPorts    map[string]interface{}
	PortBindings    map[string]types.PortBinding

	NetworkConfig  *types.NetworkSettings
	EndpointConfig *types.EndpointSettings
}
