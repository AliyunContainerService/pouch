package ctrd

import "testing"

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
	// TODO
}

func TestUseProxy(t *testing.T) {
	// TODO
}
