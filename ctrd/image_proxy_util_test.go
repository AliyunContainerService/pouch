package ctrd

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasPort(t *testing.T) {
	// TODO
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
			expected: "host:port",
		},
		{
			input:    u2,
			expected: "host:port",
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
