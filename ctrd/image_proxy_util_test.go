package ctrd

import "testing"

func TestHasPort(t *testing.T) {
	// TODO
}

func TestCanonicalAddr(t *testing.T) {
	url := "http://localhost:9090/api/download?fileidstr=2d905c2162b50d0e12b2213323b34bd4.xlsx&iswindows=0&optuser=test"

	addr := canonicalAddr(url)

	if addr == nil {
		t.Fatal("returned value is expected,but is nil")
	}
	if addr != "localhost:9090" {
		t.Fatal("Wrong return:expected localhost:9090,actual ", addr)
	}

	url = "http://localhost/api/download?fileidstr=2d905c2162b50d0e12b2213323b34bd4.xlsx&iswindows=0&optuser=test"
	addr = canonicalAddr(url)

	if addr == nil {
		t.Fatal(returned value is expected,but is nil)
	}
	if addr != "localhost:80" {
		t.Fatal("Wrong return:expected localhost:9090,actual ", addr)
	}

	url = "https://localhost/api/download?fileidstr=2d905c2162b50d0e12b2213323b34bd4.xlsx&iswindows=0&optuser=test"
	addr = canonicalAddr(url)

	if addr == nil {
		t.Fatal(returned value is expected,but is nil)
	}
	if addr != "localhost:443" {
		t.Fatal("Wrong return:expected localhost:9090,actual ", addr)
	}

	url = "socks5://localhost/api/download?fileidstr=2d905c2162b50d0e12b2213323b34bd4.xlsx&iswindows=0&optuser=test"
	addr = canonicalAddr(url)

	if addr == nil {
		t.Fatal(returned value is expected,but is nil)
	}
	if addr != "localhost:1080" {
		t.Fatal("Wrong return:expected localhost:9090,actual ", addr)
	}
}

func TestUseProxy(t *testing.T) {
	// TODO
}
