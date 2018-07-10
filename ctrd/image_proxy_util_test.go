package ctrd

import (
	"net/url"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func buildURL(inputUrl string) *url.URL {
	url, _ := url.Parse(inputUrl)
	return url
}

func TestCanonicalAddr(t *testing.T) {
	type TestCase struct {
		url      *url.URL
		expected string
	}

	testCases := []TestCase{
		{
			url:      buildURL("http://www.alibaba-inc.com"),
			expected: "www.alibaba-inc.com:80",
		},
		{
			url:      buildURL("https://www.alibaba-inc.com"),
			expected: "www.alibaba-inc.com:443",
		},
		{
			url:      buildURL("socks5://www.alibaba-inc.com"),
			expected: "www.alibaba-inc.com:1080",
		},
		{
			url:      buildURL("http://www.alibaba-inc.com:2333"),
			expected: "www.alibaba-inc.com:2333",
		},
		{
			url:      buildURL("https://www.alibaba-inc.com:2333"),
			expected: "www.alibaba-inc.com:2333",
		},
		{
			url:      buildURL("socks5://www.alibaba-inc.com:2333"),
			expected: "www.alibaba-inc.com:2333",
		},
	}

	for _, testCase := range testCases {
		addr := canonicalAddr(testCase.url)
		assert.Equal(t, testCase.expected, addr)
	}
}

func TestHasPort(t *testing.T) {
	type TestCase struct {
		input    string
		expected bool
	}

	testCases := []TestCase{
		{
			input:    "www.alibaba.com:80",
			expected: true,
		},
		{
			input:    "www.alibaba.com",
			expected: false,
		},
		{
			input:    "[www.alibab.com::80]:80",
			expected: true,
		},
		{
			input:    "[www.alibab.com::80]",
			expected: false,
		},
	}

	for _, testCase := range testCases {
		outputbool := hasPort(testCase.input)
		assert.Equal(t, testCase.expected, outputbool)
	}
}

var noproxy string

type mockNoProxyEnvGet struct {
	eo *envOnce
}

func newMockNoProxyEnvGet(e *envOnce, nopro string) *envOnce {
	noproxy = nopro
	return (*envOnce)(unsafe.Pointer(&mockNoProxyEnvGet{e}))
}

func (m *mockNoProxyEnvGet) Get() string {
	return noproxy
}

func TestuseProxy(t *testing.T) {
	type TestCase struct {
		input    string
		expected bool
		noProxy  string
	}

	testCases := []TestCase{
		{
			input:    "",
			expected: true,
			noProxy:  ".foo.com,foo.com",
		},
		{
			input:    "http://www.localhost.com",
			expected: false,
			noProxy:  "",
		},
		{
			input:    "http://www.localhost.com:8000",
			expected: false,
			noProxy:  "",
		},
		{
			input:    "http://www.127.0.0.1.com",
			expected: false,
			noProxy:  "",
		},
		{
			input:    "http://www.127.0.0.1.com:2333",
			expected: false,
			noProxy:  "",
		},
		{
			input:    "http://www.alibaba-inc.com:2333",
			expected: false,
			noProxy:  "*",
		},
		{
			input:    "http://www.alibaba-inc.com",
			expected: false,
			noProxy:  "*",
		},
		{
			input:    "http://www.alibaba-inc.com",
			expected: true,
			noProxy:  "   ,   ",
		},
		{
			input:    "http://www.alibaba-inc.com:2333",
			expected: false,
			noProxy:  "alibaba-inc.com",
		},
		{
			input:    "http://www.alibaba-inc.com:2333",
			expected: true,
			noProxy:  ":2333,:8000",
		},
		{
			input:    "http://bar.foo.com",
			expected: false,
			noProxy:  ".foo.com",
		},
		{
			input:    "http://bar.foo.com",
			expected: false,
			noProxy:  "foo.com",
		},
		{
			input:    "http://bar.bar.com",
			expected: true,
			noProxy:  "foo.com",
		},
	}

	for _, testCase := range testCases {
		noProxyEnv = newMockNoProxyEnvGet(&envOnce{
			names: []string{"NO_PROXY", "no_proxy"},
		}, testCase.noProxy)
		outputbool := useProxy(testCase.input)
		assert.Equal(t, testCase.expected, outputbool)
	}
}
