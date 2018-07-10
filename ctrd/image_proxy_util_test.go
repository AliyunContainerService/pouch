package ctrd

import "testing"

func TestHasPort(t *testing.T) {
	s := "172.0.0.1"
	result := hasPort(s)
	if result == true {
		t.Fatalf("Wrong return:expected false,actual %v", result)
	}

	s = "172.0.0.1:22"
	result = hasPort(s)
	if result == false {
		t.Fatalf("Wrong return:expected true,actual %v", result)
	}

	s = "[ipv6:172.0.0.1]:22"
	result = hasPort(s)
	if result == false {
		t.Fatalf("Wrong return:expected true,actual %v", result)
	}

	s = "[ipv6:172.0.0.1]"
	result = hasPort(s)
	if result == true {
		t.Fatalf("Wrong return:expected false,actual %v", result)
	}
}

func TestCanonicalAddr(t *testing.T) {
	// TODO

}

func TestUseProxy(t *testing.T) {
	// TODO
}
