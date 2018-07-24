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
	var tests = []struct {
		name    string
		addr    string
		noProxy string
		want    bool
	}{
		{name: "test1", addr: "127.0.0.1:8000", want: false},
		{name: "test2", addr: "10.0.0.1:8000", want: true},
		{name: "test3", addr: "30.45.23.52:77", want: true},
		{name: "test4", addr: "localhost:80", want: false},
		{name: "test5", addr: "[3ffe:ffff:7654:feda:1234:ba98:2312:2134]:80", want: true},
		{name: "test6", addr: "", want: true},
		{name: "test7", addr: "123:123", want: true},
		{name: "test8", addr: "44.45.23.52", want: false},
		{name: "test9", addr: "baidu.com", want: false},
		{name: "test10", addr: "baidu.com", noProxy: "baidu.com", want: false},
		{name: "test11", addr: "30.45.23.51:77", noProxy: "foo.com", want: true},
		{name: "test12", addr: "30.45.23.51:77", noProxy: "*", want: false},
	}

	noProxyEnv.once.Do(func() {})
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			noProxyEnv.val = testCase.noProxy
			got := useProxy(testCase.addr)
			if got != testCase.want {
				t.Errorf("useProxy() = %v, want %v", got, testCase.want)
			}
		})
	}
}
