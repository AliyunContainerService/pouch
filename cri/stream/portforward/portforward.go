package portforward

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/alibaba/pouch/pkg/log"
)

// PortForwarder knows how to forward content from a data stream to/from a port
// in a pod.
type PortForwarder interface {
	// PortForwarder copies data between a data stream and a port in a pod.
	PortForward(ctx context.Context, name string, port int32, stream io.ReadWriteCloser) error
}

// ServePortForward handles a port forwarding request. A single request is
// kept alive as long as the client is still alive and the connection has not
// been timed out due to idleness. This function handles multiple forwarded
// connections; i.e., multiple `curl http://localhost:8888/` requests will be
// handled by a single invocation of ServePortForward.
func ServePortForward(ctx context.Context, w http.ResponseWriter, req *http.Request, portForwarder PortForwarder, podName string, idleTimeout time.Duration, streamCreationTimeout time.Duration, supportedProtocols []string) {
	// TODO: support web socket stream.
	err := handleHTTPStreams(ctx, w, req, portForwarder, podName, idleTimeout, streamCreationTimeout, supportedProtocols)
	if err != nil {
		log.With(ctx).Errorf("failed to serve port forward: %v", err)
		return
	}
}
