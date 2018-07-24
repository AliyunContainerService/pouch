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
	// TODO
	type args struct {
		str string
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		group int
	}{
		// TODO: Add test cases.
		//error
		{name: "test0", args: args{str: string("[localhost")}, want: false, group: 1},
		{name: "test1", args: args{str: string("ipv6::localhost]")}, want: false, group: 1},
		{name: "test2", args: args{str: string("[ipv6::localhost][8000")}, want: false, group: 1},
		{name: "test3", args: args{str: string("localhost")}, want: false, group: 1},

		//loop back
		{name: "test4", args: args{str: string("ipv6::localhost")}, want: false, group: 1},
		{name: "test5", args: args{str: string("ipv6::localhost8000")}, want: false, group: 1},
		{name: "test6", args: args{str: string("localhost:8000")}, want: false, group: 1},
		{name: "test7", args: args{str: string("127.0.0.1:8000")}, want: false, group: 1},

		{name: "test8", args: args{str: string("196.1.1.1:4000")}, want: true, group: 1},
		{name: "test9", args: args{str: string("[ipv6::gfff:ffff:ffff:ffff:ffff:ffff:ffff:ffff]:2000")}, want: true, group: 1},

		{name: "test10", args: args{str: string("196.1.1.1:4000")}, want: false, group: 2},
		{name: "test11", args: args{str: string("[ipv6::gfff:ffff:ffff:ffff:ffff:ffff:ffff:ffff]:2000")}, want: false, group: 2},

		{name: "test12", args: args{str: string("abc.foo.com:1000")}, want: false, group: 3},
		{name: "test12", args: args{str: string("www.abc.foo.com:4000")}, want: false, group: 3},
		{name: "test12", args: args{str: string(".foo.com:4000")}, want: false, group: 3},
		{name: "test13", args: args{str: string("196.168.2.2:4000")}, want: false, group: 3},
		{name: "test14", args: args{str: string("[ipv6::1234:1234:1234:1234:1234:1234:1234:1234]:1000")}, want: true, group: 3},
		{name: "test15", args: args{str: string("196.168.5.5:4000")}, want: true, group: 3},
	}

	noProxyEnv.once.Do(func() {})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.group == 3 {
				noProxyEnv.val = "foo.com, .foo.com, ,196.168.2.2"
			} else if tt.group == 2 {
				noProxyEnv.val = "*"
			} else {
				noProxyEnv.val = ""
			}
			got := useProxy(tt.args.str)
			if got != tt.want {
				t.Errorf("useProxy() = %v, want %v", got, tt.want)
				return
			}
		})
	}

}
