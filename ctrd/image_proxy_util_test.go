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
	tests := []struct {
		name     string
		hostPort string
		wantErr  bool
	}{
		{name: "a", hostPort: "", wantErr: true},
		{name: "b", hostPort: "cannotBeSplit", wantErr: false},
		{name: "c", hostPort: "foo.com:80", wantErr: true},
		{name: "d", hostPort: "localhost:80", wantErr: false},
		{name: "e", hostPort: "192.168.0.1:80", wantErr: true},
		{name: "f", hostPort: ".foo.com:80", wantErr: true},
		{name: "f", hostPort: ".foo.com", wantErr: false},
	}
	//run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := useProxy(tt.hostPort)
			if result != tt.wantErr {
				t.Errorf("useProxy() = %v, want %v", result, tt.wantErr)
			}
		})
	}
}
