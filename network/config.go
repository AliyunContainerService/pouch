package network

// DefaultExecRoot defines the default network execute root directory.
var DefaultExecRoot = "/var/run/pouch"

// Config defines the network configuration.
type Config struct {
	Type string

	MetaPath   string // meta store
	ExecRoot   string // exec root
	DNS        []string
	DNSOptions []string
	DNSSearch  []string

	// bridge config
	BridgeConfig BridgeConfig
}

// BridgeConfig defines the bridge network configuration.
type BridgeConfig struct {
	Name        string
	IP          string
	FixedCIDR   string
	GatewayIPv4 string
	PreferredIP string

	Mtu               int
	ICC               bool
	IPTables          bool
	IPForward         bool
	IPMasq            bool
	UserlandProxy     bool
	UserlandProxyPath string
}
