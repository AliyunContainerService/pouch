package constant

const (
	// StreamType is the name of header that specifies stream type
	StreamType = "streamType"
	// StreamTypeStdin is the value for streamType header for stdin stream
	StreamTypeStdin = "stdin"
	// StreamTypeStdout is the value for streamType header for stdout stream
	StreamTypeStdout = "stdout"
	// StreamTypeStderr is the value for streamType header for stderr stream
	StreamTypeStderr = "stderr"
	// StreamTypeData is the value for streamType header for data stream
	StreamTypeData = "data"
	// StreamTypeError is the value for streamType header for error stream
	StreamTypeError = "error"
	// StreamTypeResize is the value for streamType header for terminal resize stream
	StreamTypeResize = "resize"

	// PortHeader is the name of header that specifies the port being forwarded.
	PortHeader = "port"
	// PortForwardRequestIDHeader is the name of header that specifies a request ID
	// used to associate the error and data streams for a single forwarded connection.
	PortForwardRequestIDHeader = "requestID"
)
