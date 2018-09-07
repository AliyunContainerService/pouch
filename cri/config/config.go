package config

const (
	// K8sNamespace is the namespace we use to connect containerd when CRI is enabled.
	K8sNamespace = "k8s.io"
)

// Config defines the CRI configuration.
type Config struct {
	// Listen is the listening address which servers CRI.
	Listen string `json:"listen,omitempty"`
	// NetworkPluginBinDir is the directory in which the binaries for the plugin is kept.
	NetworkPluginBinDir string `json:"network-plugin-bin-dir,omitempty"`
	// NetworkPluginConfDir is the directory in which the admin places a CNI conf.
	NetworkPluginConfDir string `json:"network-plugin-conf-dir,omitempty"`
	// SandboxImage is the image used by sandbox container.
	SandboxImage string `json:"sandbox-image,omitempty"`
	// CriVersion is the cri version
	CriVersion string `json:"cri-version,omitempty"`
	// StreamServerPort is the port which cri stream server is listening on.
	StreamServerPort string `json:"stream-server-port,omitempty"`
	// StreamServerReusePort specify whether cri stream server share port with pouchd.
	StreamServerReusePort bool
}
