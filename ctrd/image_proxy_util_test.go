package ctrd

import "testing"

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
