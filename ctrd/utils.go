package ctrd

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// NewDefaultSpec new a template spec with default.
func NewDefaultSpec(ctx context.Context) (*specs.Spec, error) {
	return containerd.GenerateSpec(ctx, nil, nil)
}

func resolver() (remotes.Resolver, error) {
	var (
		// TODO
		username  = ""
		secret    = ""
		plainHTTP = false
		refresh   = ""
		insecure  = false
	)

	// FIXME
	_ = refresh

	options := docker.ResolverOptions{
		PlainHTTP: plainHTTP,
		Tracker:   docker.NewInMemoryTracker(),
	}
	options.Credentials = func(host string) (string, string, error) {
		// Only one host
		return username, secret, nil
	}

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
		ExpectContinueTimeout: 5 * time.Second,
	}

	options.Client = &http.Client{
		Transport: tr,
	}

	return docker.NewResolver(options), nil
}
