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
	// TODO
}
