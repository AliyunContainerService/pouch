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
