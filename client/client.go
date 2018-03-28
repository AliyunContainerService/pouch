package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	defaultHost    = "unix:///var/run/pouchd.sock"
	defaultTimeout = time.Second * 10
	// defaultVersion is the version of the current stable API
	defaultVersion = "v1.24"
)

// APIClient is a API client that performs all operations
// against a pouch server
type APIClient struct {
	proto   string // socket type
	addr    string
	baseURL string
	HTTPCli *http.Client
	// version of the server talks to
	version string
}

// TLSConfig contains information of tls which users can specify
type TLSConfig struct {
	CA           string `json:"tlscacert,omitempty"`
	Cert         string `json:"tlscert,omitempty"`
	Key          string `json:"tlskey,omitempty"`
	VerifyRemote bool
}

// NewAPIClient initializes a new API client for the given host
func NewAPIClient(host string, tls TLSConfig) (CommonAPIClient, error) {
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

	version := os.Getenv("POUCH_API_VERSION")
	if version == "" {
		version = defaultVersion
	}

	return &APIClient{
		proto:   newURL.Scheme,
		addr:    addr,
		baseURL: basePath,
		HTTPCli: httpCli,
		version: version,
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
func generateTLSConfig(host string, tls TLSConfig) *tls.Config {
	// init tls config
	if tls.Key != "" && tls.Cert != "" && !strings.HasPrefix(host, "unix://") {
		tlsCfg, err := GenTLSConfig(tls.Key, tls.Cert, tls.CA)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fail to parse tls config %v", err)
			os.Exit(1)
		}
		tlsCfg.InsecureSkipVerify = !tls.VerifyRemote

		return tlsCfg
	}
	return nil
}

func generateBaseURL(u *url.URL, tls TLSConfig) string {
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

// GetAPIPath returns the versioned request path to call the api.
// It appends the query parameters to the path if they are not empty.
func (client *APIClient) GetAPIPath(path string, query url.Values) string {
	var apiPath string
	if client.version != "" {
		v := strings.TrimPrefix(client.version, "v")
		apiPath = fmt.Sprintf("/v%s%s", v, path)
	} else {
		apiPath = path
	}

	u := url.URL{
		Path: apiPath,
	}
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}
	return u.String()
}

// UpdateClientVersion sets client version new value.
func (client *APIClient) UpdateClientVersion(v string) {
	client.version = v
}

// GenTLSConfig returns a tls config object according to inputting parameters.
func GenTLSConfig(key, cert, ca string) (*tls.Config, error) {
	tlsConfig := &tls.Config{}
	tlsCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("failed to read X509 key pair (cert: %q, key: %q): %v", cert, key, err)
	}
	tlsConfig.Certificates = []tls.Certificate{tlsCert}
	if ca == "" {
		return tlsConfig, nil
	}
	cp, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("failed to read system certificates: %v", err)
	}
	pem, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate %q: %v", ca, err)
	}
	if !cp.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("failed to append certificates from PEM file: %q", ca)
	}
	tlsConfig.ClientCAs = cp
	return tlsConfig, nil
}
