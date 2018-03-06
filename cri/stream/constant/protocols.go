package constant

const (
	// StreamProtocolV1Name represents the initial unversioned subprotocol used for remote command attachment/execution.
	StreamProtocolV1Name = "channel.k8s.io"

	// StreamProtocolV2Name is the second version of the subprotocol and resolves the issues present in the first version.
	StreamProtocolV2Name = "v2.channel.k8s.io"

	// StreamProtocolV3Name is the third version of the subprotocol and adds support for resizing container terminals.
	StreamProtocolV3Name = "v3.channel.k8s.io"

	// StreamProtocolV4Name is the 4th version of the subprotocol and adds support for exit codes.
	StreamProtocolV4Name = "v4.channel.k8s.io"

	// PortForwardProtocolV1Name is the SPDY subprotocol "portforward.k8s.io" used for port forwarding.
	PortForwardProtocolV1Name = "portforward.k8s.io"
)
