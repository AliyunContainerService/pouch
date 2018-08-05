// HTTP related code copy from src/net/http/transport.go

package ctrd

import (
	"net/http"
	"net/url"
	"os"
	"sync"

	"golang.org/x/net/http/httpproxy"
)

var proxy string

// SetImageProxy sets value of image http proxy
func SetImageProxy(p string) {
	proxy = p
}

func proxyFromEnvironment(req *http.Request) (*url.URL, error) {
	config := &httpproxy.Config{
		HTTPProxy: proxy,
		NoProxy:   noProxyEnv.Get(),
	}

	return config.ProxyFunc()(req.URL)
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
