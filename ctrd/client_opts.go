package ctrd

import "fmt"

type clientOpts struct {
	startDaemon            bool
	debugLog               bool
	rpcAddr                string
	homeDir                string
	containerdBinary       string
	grpcClientPoolCapacity int
	maxStreamsClient       int
	oomScoreAdjust         int
	defaultns              string
}

// ClientOpt allows caller to set options for containerd client.
type ClientOpt func(c *clientOpts) error

// WithStartDaemon set startDaemon flag for containerd client.
// startDaemon is a flag to decide whether start a new containerd instance
// when create a containerd client.
func WithStartDaemon(startDaemon bool) ClientOpt {
	return func(c *clientOpts) error {
		c.startDaemon = startDaemon
		return nil
	}
}

// WithRPCAddr set containerd listen address.
func WithRPCAddr(rpcAddr string) ClientOpt {
	return func(c *clientOpts) error {
		if rpcAddr == "" {
			return fmt.Errorf("rpc socket path is empty")
		}

		c.rpcAddr = rpcAddr
		return nil
	}
}

// WithDebugLog set debugLog flag for containerd client.
// debugLog decides containerd log level.
func WithDebugLog(debugLog bool) ClientOpt {
	return func(c *clientOpts) error {
		c.debugLog = debugLog
		return nil
	}
}

// WithHomeDir set home dir for containerd.
func WithHomeDir(homeDir string) ClientOpt {
	return func(c *clientOpts) error {
		if homeDir == "" {
			return fmt.Errorf("containerd home Dir is empty")
		}

		c.homeDir = homeDir
		return nil
	}
}

// WithContainerdBinary specifies the containerd binary path.
func WithContainerdBinary(containerdBinary string) ClientOpt {
	return func(c *clientOpts) error {
		if containerdBinary == "" {
			return fmt.Errorf("containerd binary path is empty")
		}

		c.containerdBinary = containerdBinary
		return nil
	}
}

// WithGrpcClientPoolCapacity sets containerd clients pool capacity.
func WithGrpcClientPoolCapacity(grpcClientPoolCapacity int) ClientOpt {
	return func(c *clientOpts) error {
		if grpcClientPoolCapacity <= 0 {
			return fmt.Errorf("containerd clients pool capacity should positive number")
		}

		c.grpcClientPoolCapacity = grpcClientPoolCapacity
		return nil
	}
}

// WithMaxStreamsClient sets one containerd grpc client can hold max streams client.
func WithMaxStreamsClient(maxStreamsClient int) ClientOpt {
	return func(c *clientOpts) error {

		if maxStreamsClient <= 0 {
			return fmt.Errorf("containerd max streams client should be positive number")
		}

		c.maxStreamsClient = maxStreamsClient
		return nil
	}
}

// WithOOMScoreAdjust sets oom-score for containerd instance.
func WithOOMScoreAdjust(oomScore int) ClientOpt {
	return func(c *clientOpts) error {
		if oomScore > 1000 || oomScore < -1000 {
			return fmt.Errorf("oom-score range should be [-1000, 1000]")
		}

		c.oomScoreAdjust = oomScore
		return nil
	}
}

// WithDefaultNamespace sets the default namespace on the client
//
// Any operation that does not have a namespace set on the context will
// be provided the default namespace
func WithDefaultNamespace(ns string) ClientOpt {
	return func(c *clientOpts) error {
		c.defaultns = ns
		return nil
	}
}
