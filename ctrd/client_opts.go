package ctrd

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type clientOpts struct {
	rpcAddr                string
	grpcClientPoolCapacity int
	maxStreamsClient       int
	defaultns              string
	insecureRegistries     []string
}

// ClientOpt allows caller to set options for containerd client.
type ClientOpt func(c *clientOpts) error

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

// WithInsecureRegistries sets the insecure registries to allow http request
// and skip secure verify.
func WithInsecureRegistries(endpoints []string) ClientOpt {
	return func(c *clientOpts) error {
		registries := make([]string, 0, len(endpoints))

		for _, r := range endpoints {
			if strings.Contains(strings.ToLower(r), "://") {
				return fmt.Errorf("insecure registry %s should not contain any '://'", r)
			}

			if err := validateHostPort(r); err != nil {
				return err
			}
			registries = append(registries, r)
		}
		c.insecureRegistries = registries
		return nil
	}
}

func validateHostPort(s string) error {
	_, port, err := net.SplitHostPort(s)
	if err != nil {
		port = ""
	}

	if port != "" {
		v, err := strconv.Atoi(port)
		if err != nil {
			return err
		}

		if v < 0 || v > 65535 {
			return fmt.Errorf("invalid port %q", port)
		}
	}
	return nil
}
