package ctrd

import (
	"net/url"
	"testing"
)

func TestHasPort(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{str: string("localhost:8000")}, want: true},
		{name: "test2", args: args{str: string("[ipv6::localhost]:8000")}, want: true},
		{name: "test3", args: args{str: string(":8000")}, want: true},
		{name: "test4", args: args{str: string("[ipv6::127.0.0.1]::8000")}, want: true},
		{name: "test5", args: args{str: string("localhost")}, want: false},
		{name: "test6", args: args{str: string("[ipv6::localhost]")}, want: false},
		{name: "test7", args: args{str: string("[ipv6::localhost]8000")}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasPort(tt.args.str)
			if got != tt.want {
				t.Errorf("hasPort() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}

func TestCanonicalAddr(t *testing.T) {
	tests := []struct {
		name      string
		urlString string
		want      string
	}{
		{name: "test1", urlString: "http://google.com", want: "google.com:80"},
		{name: "test2", urlString: "https://www.github.com", want: "www.github.com:443"},
		{name: "test3", urlString: "socks5://shadowsocks.com", want: "shadowsocks.com:1080"},
		{name: "test4", urlString: "ftp://127.0.0.1", want: "127.0.0.1:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exampleURL, _ := url.Parse(tt.urlString)
			got := canonicalAddr(exampleURL)

			if got != tt.want {
				t.Errorf("canonicalAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUseProxy(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		noProxy string
		want    bool
	}{
		{name: "test1", addr: "", want: true},
		{name: "test2", addr: "google.com", want: false},
		{name: "test3", addr: "localhost:80", want: false},
		{name: "test4", addr: "127.0.0.1:80", want: false},
		{name: "test5", addr: "10.0.0.1:80", want: true},
		{name: "test6", addr: "golang.org:80", noProxy: "*", want: false},
		{name: "test7", addr: "maps.google.com:443", noProxy: ".google.com", want: false},
		{name: "test8", addr: "google.com:443", noProxy: ".google.com", want: false},
		{name: "test9", addr: "maps.google.com:443", noProxy: "google.com", want: false},
		{name: "test10", addr: "google.com:443", noProxy: "google.com", want: false},
		{name: "test11", addr: "maps.google.com:443", want: true},
	}

	// simulate the fetch of environmental variable no_proxy
	noProxyEnv.once.Do(func() {})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			noProxyEnv.val = tt.noProxy
			got := useProxy(tt.addr)

			if got != tt.want {
				t.Errorf("useProxy() = %v, want %v", got, tt.want)
			}
		})
	}
}
