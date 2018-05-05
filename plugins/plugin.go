package plugins

import (
	"sync"
)

const (
	// HandShakePath is the path for handshaking.
	HandShakePath = "/Plugin.Activate"
)

// TLSConfig contains tls info which users can specify.
type TLSConfig struct {
	CAFile             string `json:"CAFile"`
	CertFile           string `json:"CertFile"`
	KeyFile            string `json:"KeyFile"`
	InsecureSkipVerify bool   `json:"InsecureSkipVerify"`
}

// Plugin includes the Name, Addr, TLSConfig.
type Plugin struct {
	sync.Mutex

	// Name is the plugin name.
	Name string `json:"Name"`

	// Addr is the plugin address.
	Addr string `json:"Addr"`

	// TLSConfig is the tls config.
	TLSConfig *TLSConfig `json:"TLSConfig"`

	// Implements is the plugin implement infos.
	Implements []string `json:"-"`

	// Client is the plugin client.
	client *PluginClient

	// probed represents the plugin is whether probed.
	probed bool

	// probeError is the probe error.
	probeError error
}

// HandShakeResp is a handshake response.
type HandShakeResp struct {
	Implements []string `json:"Implements"`
}

// Client returns the plugin client
func (p *Plugin) Client() *PluginClient {
	return p.client
}

// isProbed checks whether the plugin is probed.
func (p *Plugin) isProbed() bool {
	return p.probed
}

// implement checks whether implement the given pluginType.
func (p *Plugin) implement(pluginType string) bool {
	for _, implement := range p.Implements {
		if implement == pluginType {
			return true
		}
	}
	return false
}

// probe will try to probe the plugin.
func (p *Plugin) probe() error {
	p.Lock()
	defer p.Unlock()

	if p.isProbed() {
		return p.probeError
	}
	return p.handshake()
}

// handshake sends handshake request to the plugin server.
func (p *Plugin) handshake() error {
	out := new(HandShakeResp)

	err := p.client.CallService(HandShakePath, nil, out, true)

	p.probed = true

	if err != nil {
		p.probeError = err
		return err
	}

	p.probeError = nil
	p.Implements = out.Implements
	return nil
}
