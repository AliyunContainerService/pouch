package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

var (
	defaultHost    = "unix:///var/run/pouchd.sock"
	defaultTimeout = time.Second * 10
)

// Client is a API client that performs all operations
// against a pouch server
type Client struct {
	proto   string // socket type
	baseURL string

	httpCli *http.Client
}

// New initializes a new API client for the given host
func New(host string) (*Client, error) {
	if host == "" {
		host = defaultHost
	}

	u, basePath, err := parseHost(host)
	if err != nil {
		return nil, fmt.Errorf("failed to new client: %s", err)
	}

	httpCli := newHTTPClient(u)

	return &Client{
		proto:   u.Scheme,
		baseURL: basePath,
		httpCli: httpCli,
	}, nil
}

func parseHost(host string) (*url.URL, string, error) {
	u, err := url.Parse(host)
	if err != nil {
		return nil, "", err
	}

	var basePath string
	switch u.Scheme {
	case "unix":
		basePath = "http://d"
	case "tcp":
		basePath = "http://" + u.Host
	case "http":
		basePath = host
	default:
		return nil, "", fmt.Errorf("not support url scheme %v", u.Scheme)
	}

	return u, basePath, nil
}

func newHTTPClient(u *url.URL) *http.Client {
	tr := &http.Transport{}

	switch u.Scheme {
	case "unix":
		unixDial := func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.DialTimeout("unix", u.Path, time.Duration(defaultTimeout))
		}
		tr.DialContext = unixDial
	default:
		dial := func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, time.Duration(defaultTimeout))
		}
		tr.DialContext = dial
	}

	return &http.Client{
		Transport: tr,
	}
}
