package ctrd

import (
	"testing"
	"net/url"

	"github.com/stretchr/testify/assert"
)
func TestHasPort(t *testing.T) {
	// mock data
	var tests = [] struct {
		s string
		expect bool
	}{
		{"", false},
		{":",true},
		{"]",false},
		{"test1", false},
		{"test2:]", false},
		{"test3]:", true},
		{"test4:]:", true},
		{"test5:]:]", false},
	}

	for _, tt := range tests {
		actual := hasPort(tt.s)
		
		// compare the result
		assert.Equal(t, tt.expect, actual)
	}
}

func TestCanonicalAddr(t *testing.T) {
	// mock data
	var tests = [] struct{
		url *url.URL
		expect string
	}{
		{
			url : &url.URL{Scheme : "sss" , Host : "127.0.0.1"},
			expect: "127.0.0.1:",
		},
	}

	for _, tt := range tests {
		actual := canonicalAddr(tt.url)
		
		// compare the result
		assert.Equal(t, tt.expect, actual)
	}
}

func TestUseProxy(t *testing.T) {
	// mock data
	var tests = [] struct{
		addr string
		expect bool
	}{
		{addr : "" , expect : true},
		{addr : "@#$%" , expect : false},
		{addr : "localhost" , expect : false},

		// need to mock system env
		{addr : "http://www.baidu.com", expect : true},
	}

	for _, tt := range tests {
		actual := useProxy(tt.addr)
		
		// compare the result
		assert.Equal(t, tt.expect, actual)
	}
}
