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
	// TODO
}

func TestUseProxy(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// empty addr, return true
		{name: "test1", args: args{str: string("")}, want: true},
		// SplitHostPort err: missing port in address, return false
		{name: "test2", args: args{str: string("123456")}, want: false},
		// missing port in address
		{name: "test3", args: args{str: string("localhost")}, want: false},
		{name: "test4", args: args{str: string("202.117.16.188")}, want: false},
		// localhost, return false
		{name: "test5", args: args{str: string("localhost:8000")}, want: false},
		// 127 IsLoopback, return false
		{name: "test6", args: args{str: string("127.0.0.1:8000")}, want: false},
		// use proxy
		{name: "test7", args: args{str: string("202.117.16.188:80")}, want: true},
		{name: "test8", args: args{str: string("202.117.16.188:12200")}, want: true},
		{name: "test9", args: args{str: string("192.168.16.188:80")}, want: true},
		{name: "test9", args: args{str: string("192.168.16.188:12200")}, want: true},
	}

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
