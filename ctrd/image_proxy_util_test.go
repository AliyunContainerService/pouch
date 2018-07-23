package ctrd

import (
	"os"
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
	// TODO
}

func TestUseProxy(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want bool
	}{
		{name: "test1", addr: "", want: true},
		{name: "test2", addr: "localhost", want: false},
		{name: "test3", addr: "localhost:8080", want: false},
		{name: "test4", addr: "127.0.0.1", want: false},
		{name: "test5", addr: "127.0.0.1:22", want: false},
		{name: "test6", addr: "100.100.100.100", want: false},
		{name: "test7", addr: "110.110.110.110:80", want: false},
		{name: "test8", addr: "120.120.120.120:80", want: false},
		{name: "test9", addr: "123.123.123.123", want: false},
		{name: "test10", addr: "240.240.240.240", want: true},
		{name: "test11", addr: "www.baidu.com", want: false},
		{name: "test12", addr: "test.www.baidu.com", want: false},
		{name: "test13", addr: "bar.foo.com", want: false},
		{name: "test14", addr: "foo.com", want: false},
		{name: "test15", addr: "www.baidu.com", want: false},
		{name: "test16", addr: "www.baidu.com", want: false},
		{name: "test17", addr: "www.baidu.com", want: false},
	}

	os.Setenv("no_proxy", "100.100.100.100,110.110.110.110:80,120.120.120.120:90,123.123.123.123,:8080,www.baidu.com,.foo.com")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := useProxy(tt.addr)
			if got != tt.want {
				t.Errorf("useProxy() = %v, want %v", got, tt.want)
				return
			}
		})
	}

	os.Setenv("NO_PROXY", "*")
	tests2 := []struct {
		name string
		addr string
		want bool
	}{
		{name: "test21", addr: "localhost", want: false},
		{name: "test22", addr: "localhost:8080", want: false},
		{name: "test23", addr: "127.0.0.1", want: false},
		{name: "test24", addr: "100.100.100.100", want: false},
		{name: "test25", addr: "110.110.110.110:80", want: false},
		{name: "test26", addr: "www.baidu.com", want: false},
		{name: "test27", addr: "test.www.baidu.com", want: false},
		{name: "test28", addr: "bar.foo.com", want: false},
		{name: "test29", addr: "foo.com", want: false},
	}

	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			got := useProxy(tt.addr)
			if got != tt.want {
				t.Errorf("useProxy() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}
