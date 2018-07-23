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
	var str = "://alibaba-inc@host.com"

	methods := []string{
		"http",
		"https",
		"socks5",
		"other",
	}

	for i := 0; i < len(methods); i++ {

		testUrl := methods[i] + str

		parseUrl, err := url.Parse(testUrl)

		if err != nil {
			t.Fatal(err)
		}

		addr := canonicalAddr(parseUrl)

		if hasPort(addr) == false {
			t.Fatalf("no port!!")
		}
	}
}

func TestUseProxy(t *testing.T) {
	// TODO
}
