package ctrd

import (
	"testing"
	"net/url"
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
	//var userinfo *Userinfo
	//userinfo = &Userinfo{"username","password",true}
	var testUrl *url.URL
	testUrl = &url.URL{"scheme","opaque",nil,
		"127.0.0.1:8080","/usr/twc","rawPath",true,"rawQuery","fragment"};

	var result = canonicalAddr(testUrl)

	if result != testUrl.Host {
		t.Fatal("the result is error!")
	}
	return ;
}

func TestUseProxy(t *testing.T) {
	// TODO
}
