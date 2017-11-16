package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/alibaba/pouch/pkg/utils"
)

var (
	defaultHost    = "unix:///var/run/pouchd.sock"
	defaultTimeout = time.Second * 10
)

// APIClient is a API client that performs all operations
// against a pouch server
type APIClient struct {
	proto   string // socket type
	addr    string
	baseURL string
	HTTPCli *http.Client
}

// NewAPIClient initializes a new API client for the given host
func NewAPIClient(host string, tls utils.TLSConfig) (*APIClient, error) {
	if host == "" {
		host = defaultHost
	}

	newURL, basePath, addr, err := parseHost(host)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host %s: %v", host, err)
	}

	tlsConfig := generateTLSConfig(host, tls)

	httpCli := newHTTPClient(newURL, tlsConfig)

	basePath = generateBaseURL(newURL, tls)

	return &APIClient{
		proto:   newURL.Scheme,
		addr:    addr,
		baseURL: basePath,
		HTTPCli: httpCli,
	}, nil
}

// parseHost inputs a host address string, and output three type:
// url.URL, basePath and an error
func parseHost(host string) (*url.URL, string, string, error) {
	u, err := url.Parse(host)
	if err != nil {
		return nil, "", "", err
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
		return nil, "", "", fmt.Errorf("not support url scheme %v", u.Scheme)
	}

	return u, basePath, strings.TrimPrefix(host, u.Scheme+"://"), nil
}

func newHTTPClient(u *url.URL, tlsConfig *tls.Config) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

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

// generateTLSConfig configures TLS for API Client.
func generateTLSConfig(host string, tls utils.TLSConfig) *tls.Config {
	// init tls config
	if tls.Key != "" && tls.Cert != "" && !strings.HasPrefix(host, "unix://") {
		tlsCfg, err := utils.GenTLSConfig(tls.Key, tls.Cert, tls.CA)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fail to parse tls config %v", err)
			os.Exit(1)
		}
		tlsCfg.InsecureSkipVerify = !tls.VerifyRemote

		return tlsCfg
	}
	return nil
}

func generateBaseURL(u *url.URL, tls utils.TLSConfig) string {
	if tls.Key != "" && tls.Cert != "" && u.Scheme != "unix" {
		return "https://" + u.Host
	}

	if u.Scheme == "unix" {
		return "http://d"
	}
	return "http://" + u.Host
}

// BaseURL returns the base URL of APIClient
func (client *APIClient) BaseURL() string {
	return client.baseURL
}
