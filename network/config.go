package network

// DefaultExecRoot defines the default network execute root directory.
var DefaultExecRoot = "/var/run/pouch"

// Config defines the network configuration.
type Config struct {
	Type string `json:"-"`

	MetaPath   string   `json:"meta-path"`     // meta store
	ExecRoot   string   `json:"exec-root-dir"` // exec root
	DNS        []string `json:"dns"`
	DNSOptions []string `json:"dns-options"`
	DNSSearch  []string `json:"dns-search"`

	// bridge config
	BridgeConfig BridgeConfig `json:"bridge-config"`

	ActiveSandboxes map[string]interface{} `json:"-"`
}

// BridgeConfig defines the bridge network configuration.
type BridgeConfig struct {
	Name        string `json:"bridge-name"`
	IP          string `json:"bip"`
	FixedCIDR   string `json:"fixed-cidr"`
	GatewayIPv4 string `json:"default-gateway"`
	PreferredIP string `json:"preferred-ip"`

	Mtu           int  `json:"mtu"`
	ICC           bool `json:"icc"`
	IPTables      bool `json:"iptables"`
	IPForward     bool `json:"ipforward"`
	IPMasq        bool `json:"ipmasq"`
	UserlandProxy bool `json:"userland-proxy"`
}
