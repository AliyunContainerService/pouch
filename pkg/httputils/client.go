package httputils

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ParseHost inputs a host address string, and output four type:
// url.URL, basePath, address without scheme and an error.
func ParseHost(host string) (*url.URL, string, string, error) {
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
	case "https":
		basePath = host
	default:
		return nil, "", "", fmt.Errorf("not support url scheme %v", u.Scheme)
	}

	return u, basePath, strings.TrimPrefix(host, u.Scheme+"://"), nil
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

	cp := x509.NewCertPool()
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

// NewHTTPClient creates a http client using url and tlsconfig
func NewHTTPClient(u *url.URL, tlsConfig *tls.Config, dialTimeout, cliTimeout time.Duration) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	switch u.Scheme {
	case "unix":
		unixDial := func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.DialTimeout("unix", u.Path, dialTimeout)
		}
		tr.DialContext = unixDial
	default:
		dial := func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, dialTimeout)
		}
		tr.DialContext = dial
	}

	cli := &http.Client{
		Transport: tr,
	}
	if cliTimeout != 0 {
		cli.Timeout = cliTimeout
	}
	return cli
}
