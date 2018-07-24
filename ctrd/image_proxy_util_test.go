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
	// TODO

	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		/* test for address that length is zreo */
		{name: "test1", args: args{str: string("")}, want: true},
		/* test for invalid ip address */
		{name: "test2", args: args{str: string("123456")}, want: false},
		/* test for localhost */
		{name: "test3", args: args{str: string("localhost:8080")}, want: false},
		/* test for loopback address */
		{name: "test4", args: args{str: string("127.0.0.1:8080")}, want: false},
		/* test for normal address */
		{name: "test5", args: args{str: string("alibaba.com")}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := useProxy(tt.args.str)
			if got != tt.want {
				t.Errorf("hasPort() = %v, want %v", got, tt.want)
				return
			}
		})
	}

}
