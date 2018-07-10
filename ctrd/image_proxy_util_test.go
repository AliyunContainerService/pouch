package ctrd

import (
	"net/url"
	"testing"
)

func TestHasPort(t *testing.T) {
	// TODO
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

