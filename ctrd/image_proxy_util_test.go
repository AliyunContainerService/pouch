package ctrd

import (
	"testing"
	"fmt"
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

var (
	noProxyEnv = &envOnce{
		names: []string{"NO_PROXY", "no_proxy"},
	}
)

func TestUseProxy(t *testing.T) {
	// TODO
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{str: string("")}, want: true},
		{name: "test2", args: args{str: string("127.0.0.1")}, want: false},
		{name: "test3", args: args{str: string("[127.0.0.1")}, want: false},
		{name: "test4", args: args{str: string("[127.0.0.1]")}, want: false},
		{name: "test5", args: args{str: string("[127.0.0.1]8080")}, want: false},
		{name: "test6", args: args{str: string("[127.0.0.1]::8080")}, want: false},
		{name: "test7", args: args{str: string("127.0.0.1::123")}, want: false},
		{name: "test8", args: args{str: string("[[127.0.0.1]:8080")}, want: false},
		{name: "test9", args: args{str: string("[127.0.0.1]]:8080")}, want: false},
		{name: "test10", args: args{str: string("localhost:8080")}, want: false},
		{name: "test11", args: args{str: string("localhost:8080")}, want: false},
		{name: "test12", args: args{str: string("localhost:8080")}, want: false},
	}

	noProxy := noProxyEnv.Get()

	fmt.Println(noProxy)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := useProxy(tt.args.str)
			if got != tt.want {
				t.Errorf("useProxy() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}
