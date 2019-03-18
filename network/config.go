package network

var (
	// DefaultExecRoot defines the default network execute root directory.
	DefaultExecRoot = "/var/run/pouch"
	// DefaultNetworkMtu is the default value for network MTU
	DefaultNetworkMtu = 1500
)

// Config defines the network configuration.
type Config struct {
	Type string `json:"-"`

	MetaPath   string   `json:"meta-path,omitempty"`     // meta store
	ExecRoot   string   `json:"exec-root-dir,omitempty"` // exec root
	DNS        []string `json:"dns,omitempty"`
	DNSOptions []string `json:"dns-options,omitempty"`
	DNSSearch  []string `json:"dns-search,omitempty"`

	// bridge config
	BridgeConfig BridgeConfig `json:"bridge-config,omitempty"`

	ActiveSandboxes map[string]interface{} `json:"-"`
}

// BridgeConfig defines the bridge network configuration.
type BridgeConfig struct {
	DisableBridge bool   `json:"disable-bridge,omitempty"`
	Name          string `json:"bridge-name,omitempty"`
	IPv4          string `json:"bip,omitempty"`
	FixedCIDRv4   string `json:"fixed-cidr,omitempty"`
	GatewayIPv4   string `json:"default-gateway,omitempty"`
	EnableIPv6    bool   `json:"enable-ipv6,omitempty"`
	FixedCIDRv6   string `json:"fixed-cidr-v6,omitempty"`
	GatewayIPv6   string `json:"default-gateway-v6,omitempty"`
	PreferredIP   string `json:"preferred-ip,omitempty"`

	Mtu           int  `json:"mtu,omitempty"`
	ICC           bool `json:"icc,omitempty"`
	IPTables      bool `json:"iptables"`
	IPForward     bool `json:"ipforward"`
	IPMasq        bool `json:"ipmasq,omitempty"`
	UserlandProxy bool `json:"userland-proxy"`
}
