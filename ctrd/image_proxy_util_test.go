package ctrd

import (
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
	type args struct {
		str     string
		noProxy string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		// if length is 0, return true
		{name: "test1", args: args{str: string(""), noProxy: string("")}, want: true},
		// if throw an error, return false
		{name: "test2", args: args{str: string("~~~~"), noProxy: string("")}, want: false},
		// if host is "localhost", return false
		{name: "test3", args: args{str: string("localhost:"), noProxy: string("")}, want: false},
		// if address and proxy is the same, return false
		{name: "test4", args: args{str: string("taobao.com"), noProxy: string("taobao.com")}, want: false},
		// if noProxy is *, return false
		{name: "test5", args: args{str: string("localhost:8000"), noProxy: string("*")}, want: false},
		// if address is the suffix of any noProxy, return false, split by comma
		{name: "test6", args: args{str: string("taobao.com"), noProxy: string(".taobao.com, tmall.com")}, want: false},
		// if address is the suffix of any noProxy, return false
		{name: "test7", args: args{str: string("taobao.com"), noProxy: string("ali.taobao.com")}, want: false},
		// if the case does not match any above cases, return true
		{name: "test8", args: args{str: string("taobao.com:8000"), noProxy: string("")}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := useProxy(tt.args.str)
			if got != tt.want {
				t.Errorf("%v: useProxy() = %v, want %v", tt.name, got, tt.want)
				return
			}
		})
	}
}
