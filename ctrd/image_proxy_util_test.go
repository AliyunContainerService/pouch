package ctrd

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"testing"
)

func TestCanonicalAddr(t *testing.T) {
	type args struct {
		url *url.URL
	}

	var user1 = url.User("felix")
	var url1 = url.URL{Scheme: "http", Opaque: "", User: user1, Host: "127.0.0.1", Path: "/", RawPath: "", ForceQuery: false, RawQuery: "", Fragment: ""}
	var url2 = url.URL{Scheme: "https", Opaque: "", User: user1, Host: "127.0.0.1", Path: "/", RawPath: "", ForceQuery: false, RawQuery: "", Fragment: ""}
	var url3 = url.URL{Scheme: "socks5", Opaque: "", User: user1, Host: "127.0.0.1", Path: "/", RawPath: "", ForceQuery: false, RawQuery: "", Fragment: ""}
	var url4 = url.URL{Scheme: "socks5", Opaque: "", User: user1, Host: "127.0.0.1:22", Path: "/", RawPath: "", ForceQuery: false, RawQuery: "", Fragment: ""}

	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{name: "test1", args: args{url: &url1}, wantErr: "127.0.0.1:80"},
		{name: "test2", args: args{url: &url2}, wantErr: "127.0.0.1:443"},
		{name: "test3", args: args{url: &url3}, wantErr: "127.0.0.1:1080"},
		{name: "test4", args: args{url: &url4}, wantErr: "127.0.0.1:22"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := canonicalAddr(tt.args.url)
			assert.Equal(t, tt.wantErr, err)

		})
	}
}

func TestHasPort(t *testing.T) {
	type args struct {
		s string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{

		{name: "test1", args: args{s: "127.0.0.1"}, wantErr: false},
		{name: "test2", args: args{s: "127.0.0.1:80"}, wantErr: true},
		{name: "test3", args: args{s: "[ FF01::1101]:80"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hasPort(tt.args.s)
			assert.Equal(t, tt.wantErr, err)

		})
	}
}

func TestUseProxy(t *testing.T) {

	os.Setenv("NO_PROXY", "www.taobao.com:80,www.taobao1.com:80")
	noProxyEnv.init()

	type args struct {
		addr string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{

		{name: "test1", args: args{addr: ""}, wantErr: true},
		{name: "test1", args: args{addr: "localhost:80"}, wantErr: false},
		{name: "test1", args: args{addr: "127.0.0.1:80"}, wantErr: false},
		{name: "test1", args: args{addr: "www.taobao.com:80"}, wantErr: false},
		{name: "test1", args: args{addr: "www.taobao1.com:80"}, wantErr: false},
		{name: "test1", args: args{addr: "www.taobao2.com:80"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := useProxy(tt.args.addr)
			assert.Equal(t, tt.wantErr, err)

		})
	}
}
