package ctrd

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
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
	type TestCase struct {
		input    *url.URL
		expected string
	}

	s1 := "postgres://user:pass@host.com:5432/path?k=v#f"
	u1, err1 := url.Parse(s1)
	if err1 != nil {
		fmt.Println(u1)
	}

	s2 := "https://user:pass@host.com:5432/path?k=v#f"
	u2, err2 := url.Parse(s2)
	if err2 != nil {
		fmt.Println(u2)
	}

	testCases := []TestCase{
		{
			input:    u1,
			expected: "host.com:5432",
		},
		{
			input:    u2,
			expected: "host.com:5432",
		},
	}

	for _, testCase := range testCases {
		output := canonicalAddr(testCase.input)
		assert.Equal(t, testCase.expected, output)
	}
}

func TestUseProxy(t *testing.T) {
	// TODO
}
