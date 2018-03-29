package remotecommand

import (
	"fmt"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/util/wsstream"
)

const (
	stdinChannel = iota
	stdoutChannel
	stderrChannel
	errorChannel
	resizeChannel

	preV4BinaryWebsocketProtocol = wsstream.ChannelWebSocketProtocol
	preV4Base64WebsocketProtocol = wsstream.Base64ChannelWebSocketProtocol
	v4BinaryWebsocketProtocol    = "v4." + wsstream.ChannelWebSocketProtocol
	v4Base64WebsocketProtocol    = "v4." + wsstream.Base64ChannelWebSocketProtocol
)

// createChannels returns the standard channel types for a shell connection (STDIN 0, STDOUT 1, STDERR 2)
// along with the approximate duplex value. It also creates the error (3) and resize (4) channels.
func createChannels(opts *Options) []wsstream.ChannelType {
	// open the requested channels, and always open the error channel
	channels := make([]wsstream.ChannelType, 5)
	channels[stdinChannel] = readChannel(opts.Stdin)
	channels[stdoutChannel] = writeChannel(opts.Stdout)
	channels[stderrChannel] = writeChannel(opts.Stderr)
	channels[errorChannel] = wsstream.WriteChannel
	channels[resizeChannel] = wsstream.ReadChannel
	return channels
}

// readChannel returns wsstream.ReadChannel if real is true, or wsstream.IgnoreChannel.
func readChannel(real bool) wsstream.ChannelType {
	if real {
		return wsstream.ReadChannel
	}
	return wsstream.IgnoreChannel
}

// writeChannel returns wsstream.WriteChannel if real is true, or wsstream.IgnoreChannel.
func writeChannel(real bool) wsstream.ChannelType {
	if real {
		return wsstream.WriteChannel
	}
	return wsstream.IgnoreChannel
}

// createWebSocketStreams returns a context containing the websocket connection and
// streams needed to perform an exec or an attach.
func createWebSocketStreams(w http.ResponseWriter, req *http.Request, opts *Options, idleTimeout time.Duration) (*context, bool) {
	channels := createChannels(opts)
	conn := wsstream.NewConn(map[string]wsstream.ChannelProtocolConfig{
		"": {
			Binary:   true,
			Channels: channels,
		},
		preV4BinaryWebsocketProtocol: {
			Binary:   true,
			Channels: channels,
		},
		preV4Base64WebsocketProtocol: {
			Binary:   false,
			Channels: channels,
		},
		v4BinaryWebsocketProtocol: {
			Binary:   true,
			Channels: channels,
		},
		v4Base64WebsocketProtocol: {
			Binary:   false,
			Channels: channels,
		},
	})
	conn.SetIdleTimeout(idleTimeout)
	negotiatedProtocol, streams, err := conn.Open(w, req)
	if err != nil {
		runtime.HandleError(fmt.Errorf("Unable to upgrade websocket connection: %v", err))
		return nil, false
	}

	// Send an empty message to the lowest writable channel to notify the client the connection is established
	// TODO: make generic to SPDY and WebSockets and do it outside of this method?
	switch {
	case opts.Stdout:
		streams[stdoutChannel].Write([]byte{})
	case opts.Stderr:
		streams[stderrChannel].Write([]byte{})
	default:
		streams[errorChannel].Write([]byte{})
	}

	ctx := &context{
		conn:         conn,
		stdinStream:  streams[stdinChannel],
		stdoutStream: streams[stdoutChannel],
		stderrStream: streams[stderrChannel],
		resizeStream: streams[resizeChannel],
		tty:          opts.TTY,
	}

	switch negotiatedProtocol {
	case v4BinaryWebsocketProtocol, v4Base64WebsocketProtocol:
		ctx.writeStatus = v4WriteStatusFunc(streams[errorChannel])
	default:
		ctx.writeStatus = v1WriteStatusFunc(streams[errorChannel])
	}

	return ctx, true
}
