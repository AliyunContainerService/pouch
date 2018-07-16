package stream

import (
	"net/url"
	"time"

	"github.com/alibaba/pouch/cri/stream/constant"
)

// Keep these constants consistent with the peers in official package:
// k8s.io/kubernetes/pkg/kubelet/server.
const (
	// DefaultStreamIdleTimeout is the timeout for idle stream.
	DefaultStreamIdleTimeout = 4 * time.Hour

	// DefaultStreamCreationTimeout is the timeout for stream creation.
	DefaultStreamCreationTimeout = 30 * time.Second
)

// TODO: StreamProtocolV2Name, StreamProtocolV3Name, StreamProtocolV4Name support.

// SupportedStreamingProtocols is the streaming protocols which server supports.
var SupportedStreamingProtocols = []string{constant.StreamProtocolV1Name, constant.StreamProtocolV2Name}

// SupportedPortForwardProtocols is the portforward protocols which server supports.
var SupportedPortForwardProtocols = []string{constant.PortForwardProtocolV1Name}

// Config defines the options used for running the stream server.
type Config struct {
	// Address is the addr:port address the server will listen on.
	Address string

	// BaseURL is the optional base URL for constructing streaming URLs.
	// If empty, the baseURL will be constructed from the serve address.
	BaseURL *url.URL

	// StreamIdleTimeout is how long to leave idle connections open for.
	StreamIdleTimeout time.Duration
	// StreamCreationTimeout is how long to wait for clients to create streams. Only used for SPDY streaming.
	StreamCreationTimeout time.Duration

	// SupportedStreamingProtocols is the streaming protocols which server supports.
	SupportedRemoteCommandProtocols []string
	// SupportedPortForwardProtocol is the portforward protocols which server supports.
	SupportedPortForwardProtocols []string
}

// DefaultConfig provides default values for server Config.
var DefaultConfig = Config{
	StreamIdleTimeout:               DefaultStreamIdleTimeout,
	StreamCreationTimeout:           DefaultStreamCreationTimeout,
	SupportedRemoteCommandProtocols: SupportedStreamingProtocols,
	SupportedPortForwardProtocols:   SupportedPortForwardProtocols,
}
