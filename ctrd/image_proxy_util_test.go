package ctrd

import (
	"net/url"
	"testing"
)

func TestHasPort(t *testing.T) {
	// TODO
}

func TestCanonicalAddr(t *testing.T) {
	urlRawStr0 := "http://user.pass@host.com:80/path?k=v#f"
	u0, err0 := url.Parse(urlRawStr0)
	if err0 != nil {
		t.Fatal(err0)
	}
	urlCanStr0 := canonicalAddr(u0)
	has := hasPort(urlCanStr0)
	if has == false {
		t.Fatalf("canonicalAddr url %v has no port", urlRawStr0)
	}

	schemaStr := [4]string{
		"http",
		"https",
		"socks5",
		"go",
	}

	for i := 0; i < len(schemaStr); i++ {
		urlRawStr := schemaStr[i] + "://user.pass@host.com/path?k=v#f"

		u, err := url.Parse(urlRawStr)
		if err != nil {
			t.Fatal(err)
		}

		urlCanStr := canonicalAddr(u)

		has := hasPort(urlCanStr)
		if has == false {
			t.Fatalf("canonicalAddr url %v has no port", urlRawStr)
		}
	}
}

func TestUseProxy(t *testing.T) {
	// TODO
}
