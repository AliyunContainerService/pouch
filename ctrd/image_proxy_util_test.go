package ctrd

import (
	"net/url"
	"testing"
)

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

func TestHasPort(t *testing.T) {
	s := "127.0.0.1:80"
	has := hasPort(s)
	if has != true {
		t.Fatalf("expect host:port %s has port, but return false", s)
	}

	s = "[FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF::127.0.0.1]:80"
	has = hasPort(s)
	if has != true {
		t.Fatalf("expect [ipv6::address]:port %s has port, but return false", s)
	}

	s = "127.0.0.1"
	has = hasPort(s)
	if has == true {
		t.Fatalf("expect host %s has not port, but return true", s)
	}

	s = "[FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF:FFFF::127.0.0.1]"
	has = hasPort(s)
	if has == true {
		t.Fatalf("expect [ipv6::address]:port %s has not port, but return true", s)
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
