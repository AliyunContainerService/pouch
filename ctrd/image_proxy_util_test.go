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
		{name: "FormatError", args: args{str: string("352.242.123.312::8080")}, want: false},
		{name: "StringIsNull", args: args{str: string("")}, want: true},
		{name: "localhost", args: args{str: string("localhost:8000")}, want: false},
		{name: "isLoopbak", args: args{str: string("127.0.0.1:8000")}, want: false},
		{name: "ipv6Localhost", args: args{str: string("[ipv6::localhost]:8000")}, want: true},
		{name: "ipv6IsNUll", args: args{str: string("[ipv6::127.0.0.1]:8000")}, want: true},
		{name: "ipv6FormatError", args: args{str: string("[ipv6::127.0.0.1]::8000")}, want: false},
		{name: "ipv6http", args: args{str: string("[ipv6::127.0.0.1]:http")}, want: true},
		{name: "missHost", args: args{str: string(":8000")}, want: true},
		{name: "missPort", args: args{str: string("localhost")}, want: false},
		{name: "ipv6missPort", args: args{str: string("[ipv6::localhost]")}, want: false},
		{name: "ipv6missSemicolon", args: args{str: string("[ipv6::localhost]8000")}, want: false},
		{name: "normalPass", args: args{str: string("354.172.12.1:8080")}, want: true},
		{name: "matchEnv1", args: args{str: string("bar.foo.com:8080")}, want: false},
		{name: "matchEnv2", args: args{str: string("foo.com:8080")}, want: false},
		{name: "matchEnv3", args: args{str: string("bar.foo.com:8080")}, want: false},
		{name: "notMatchEnv", args: args{str: string("www.alibaba.com:8080")}, want: true},
		{name: "allFalseEnv", args: args{str: string("localhost:8080")}, want: false},
	}
	env := []struct {
		name string
		val  string
	}{
		//The environment variable for testCast
		{name: "NO_PROXY", val: ".foo.com"},
		{name: "NO_PROXY", val: ".foo.com"},
		{name: "NO_PROXY", val: ".foo.com"},
		{name: "NO_PROXY", val: ".foo.com"},
		{name: "NO_PROXY", val: ".foo.com"},
		{name: "NO_PROXY", val: ".foo.com"},
		{name: "NO_PROXY", val: ".foo.com"},
		{name: "NO_PROXY", val: ".foo.com"},
		{name: "NO_PROXY", val: ".foo.com"},
		{name: "NO_PROXY", val: ".foo.com"},
		{name: "NO_PROXY", val: ".foo.com"},
		{name: "NO_PROXY", val: ".foo.com"},
		{name: "NO_PROXY", val: ".foo.com"},
		{name: "NO_PROXY", val: ".foo.com"},
		{name: "NO_PROXY", val: ".foo.com"},
		{name: "NO_PROXY", val: "foo.com"},
		{name: "NO_PROXY", val: "foo.com,.foo.com,baidu.com"},
		{name: "NO_PROXY", val: "*"},
	}

	count := 0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(env[count].name, env[count].val)
			count = count + 1
			got := useProxy(tt.args.str)
			if got != tt.want {
				t.Errorf("useProxy() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}
