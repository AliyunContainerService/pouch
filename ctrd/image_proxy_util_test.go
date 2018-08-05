package ctrd

import (
	"net/http"
	"net/url"
	"testing"
)

func TestProxyFromEnvironment(t *testing.T) {
	SetImageProxy("http://proxy.example.com")

	tests := []struct {
		name    string
		addr    string
		noProxy string
		want    string
	}{
		{name: "test1", addr: "", want: "http://proxy.example.com"},
		{name: "test2", addr: "http://localhost", want: ""},
		{name: "test3", addr: "http://127.0.0.1", want: ""},
		{name: "test4", addr: "http://10.0.0.1", want: "http://proxy.example.com"},
		{name: "test5", addr: "http://golang.org", noProxy: "*", want: ""},
		{name: "test6", addr: "https://maps.google.com", noProxy: ".google.com", want: ""},
		{name: "test7", addr: "https://google.com", noProxy: ".google.com", want: "http://proxy.example.com"},
		{name: "test8", addr: "https://maps.google.com", noProxy: "google.com", want: ""},
		{name: "test9", addr: "https://google.com", noProxy: "google.com", want: ""},
		{name: "test10", addr: "https://maps.google.com", want: "http://proxy.example.com"},
	}

	// simulate the fetch of environmental variable no_proxy
	noProxyEnv.once.Do(func() {})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			noProxyEnv.val = tt.noProxy
			u, _ := url.Parse(tt.addr)
			req := &http.Request{
				URL: u,
			}
			got, _ := proxyFromEnvironment(req)

			if got != nil && got.String() != tt.want {
				t.Errorf("useProxy() = %v, want %v", got.String(), tt.want)
			} else if got == nil && tt.want != "" {
				t.Errorf("useProxy() = %v, want %v", got, tt.want)
			}
		})
	}
}
