// HTTP related code copy from src/net/http/transport.go

package ctrd

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

var proxy string

// SetImageProxy sets value of image http proxy
func SetImageProxy(p string) {
	proxy = p
}

func proxyFromEnvironment(req *http.Request) (*url.URL, error) {
	if proxy == "" || req.URL.Scheme == "https" {
		return nil, nil
	}
	if !useProxy(canonicalAddr(req.URL)) {
		return nil, nil
	}
	proxyURL, err := url.Parse(proxy)
	if err != nil ||
		(proxyURL.Scheme != "http" &&
			proxyURL.Scheme != "https" &&
			proxyURL.Scheme != "socks5") {
		// proxy was bogus. Try prepending "http://" to it and
		// see if that parses correctly. If not, we fall
		// through and complain about the original one.
		if proxyURL, err := url.Parse("http://" + proxy); err == nil {
			return proxyURL, nil
		}

	}
	if err != nil {
		return nil, fmt.Errorf("invalid proxy address %q: %v", proxy, err)
	}
	return proxyURL, nil
}

var (
	portMap = map[string]string{
		"http":   "80",
		"https":  "443",
		"socks5": "1080",
	}
)

// canonicalAddr returns url.Host but always with a ":port" suffix
func canonicalAddr(url *url.URL) string {
	addr := url.Hostname()
	port := url.Port()
	if port == "" {
		port = portMap[url.Scheme]
	}
	return net.JoinHostPort(addr, port)
}

// change no_proxy to noProxy for avoid fail in go lint check
var (
	noProxyEnv = &envOnce{
		names: []string{"NO_PROXY", "no_proxy"},
	}
)

// envOnce looks up an environment variable (optionally by multiple
// names) once. It mitigates expensive lookups on some platforms
// (e.g. Windows).
type envOnce struct {
	names []string
	once  sync.Once
	val   string
}

func (e *envOnce) Get() string {
	e.once.Do(e.init)
	return e.val
}

func (e *envOnce) init() {
	for _, n := range e.names {
		e.val = os.Getenv(n)
		if e.val != "" {
			return
		}
	}
}

// useProxy reports whether requests to addr should use a proxy,
// according to the NO_PROXY or noProxy environment variable.
// addr is always a canonicalAddr with a host and port.
func useProxy(addr string) bool {
	if len(addr) == 0 {
		return true
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() {
			return false
		}
	}

	noProxy := noProxyEnv.Get()
	if noProxy == "*" {
		return false
	}

	addr = strings.ToLower(strings.TrimSpace(addr))
	if hasPort(addr) {
		addr = addr[:strings.LastIndex(addr, ":")]
	}

	for _, p := range strings.Split(noProxy, ",") {
		p = strings.ToLower(strings.TrimSpace(p))
		if len(p) == 0 {
			continue
		}
		if hasPort(p) {
			p = p[:strings.LastIndex(p, ":")]
		}
		if addr == p {
			return false
		}
		if len(p) == 0 {
			// There is no host part, likely the entry is malformed; ignore.
			continue
		}
		if p[0] == '.' && (strings.HasSuffix(addr, p) || addr == p[1:]) {
			// noProxy ".foo.com" matches "bar.foo.com" or "foo.com"
			return false
		}
		if p[0] != '.' && strings.HasSuffix(addr, p) && addr[len(addr)-len(p)-1] == '.' {
			// noProxy "foo.com" matches "bar.foo.com"
			return false
		}
	}
	return true
}

// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }
