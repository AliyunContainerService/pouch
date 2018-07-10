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
	"testing"
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

func TestcanonicalAddr(t *testing.T) {
	type args struct {
		env []string
	}
	url1, err := url.Parse("http://192.168.2.11:3306")
	fmt.Print(url1, err)
	url2, _ := url.Parse("http://10.3.8.211")
	url3, _ := url.Parse("http://123.57.232.18:8080")

	tests := []struct {
		name    string
		args    *url.URL
		want    string
	}{
		{name: "test1", args: url1,  want: "192.168.2.11:3306"},
		{name: "test2", args: url2,  want: "10.3.8.211:8080"},
		{name: "test3", args: url3,  want: "123.57.232.18:8080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := canonicalAddr(tt.args)
			if(res == tt.want) {
               return
			}
		})
	}
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
func hasPort(s string) bool { 
	return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") 
}

func TesthasPort(t *testing.T) {
	type args struct {
		env []string
	}

	tests := []struct {
		name    string
		args    string
		want    bool
	}{
		{name: "test1", args: "192.168.2.170:8080",  want: true},
		{name: "test2", args: "10.3.8.211",          want: false},
		{name: "test3", args: "123.56.27.141:3306",  want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := hasPort(tt.args)
			if(res == tt.want) {
               return
			}
		})
	}
}
