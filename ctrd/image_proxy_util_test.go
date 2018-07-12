package ctrd

import (
	"testing"
	"net/url"
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
		name string
		url  url.URL
		want bool
	}{
		{url: url.URL{Scheme: string("http"), Host: string("localhost")}, want: true},
		{url: url.URL{Scheme: string("http"), Host: string("192.168.1.1")}, want: true},
		{url: url.URL{Scheme: string("http"), Host: string("127.0.0.1")}, want: true},
		{url: url.URL{Scheme: string("https"), Host: string("127.0.0.1")}, want: true},
		{url: url.URL{Scheme: string("x"), Host: string("127.0.0.1:8081")}, want: true},
	}
	portMap = map[string]string{
		"http":   "80",
		"https":  "443",
		"socks5": "1080",
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := canonicalAddr(&test.url)
			//if url.URL.Port(test.url.Host)!="" {
			//	if got != test.url.Host+":80" {
			//		t.Errorf("canonicalAddr() = %v, want %v", got, test.want)
			//		return
			//	}
			//}
			if got != test.url.Host+":"+portMap[test.url.Scheme] {
				t.Errorf("canonicalAddr() = %v, want %v", got, test.want)
				return
			}
		})
	}

}

func TestUseProxy(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want bool
	}{
		{addr: string("192.168.1.1:80"), want: true},
		{addr: string("192.168.1.2:80"), want: true},
		{addr: string("192.168.1.3:80"), want: true},
		{addr: string("192.168.1.4:80"), want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := useProxy(test.addr)
			if got == false {
				t.Errorf("useProxy(%v) = %v, want %v", test.addr, got, test.want)
				return
			}
		})
	}
}
