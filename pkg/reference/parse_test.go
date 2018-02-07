package reference

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultTagIfMissing(t *testing.T) {
	var named Named

	// only name
	named = namedReference{"pouch"}
	named = WithDefaultTagIfMissing(named)
	assert.Equal(t, true, strings.Contains(named.String(), "latest"))

	// name:tag
	named = taggedReference{
		Named: namedReference{"pouch"},
		tag:   "1.0",
	}
	named = WithDefaultTagIfMissing(named)
	assert.Equal(t, false, strings.Contains(named.String(), "latest"))
}

func TestDomain(t *testing.T) {
	type tCase struct {
		name   string
		input  string
		domain string
		ok     bool
	}

	for _, tc := range []tCase{
		{
			name:   "Normal",
			input:  "docker.io/library/nginx:alpine",
			domain: "docker.io",
			ok:     true,
		}, {
			name:   "IP Registry",
			input:  "255.255.255.255/nginx",
			domain: "255.255.255.255",
			ok:     true,
		}, {
			name:   "Localhost registry",
			input:  "localhost:80/nginx",
			domain: "localhost:80",
			ok:     true,
		}, {
			name:   "Repo and Name",
			input:  "library/nginx",
			domain: "",
			ok:     false,
		}, {
			name:   "Only Name",
			input:  "nginx",
			domain: "",
			ok:     false,
		},
	} {
		d, ok := Domain(tc.input)
		assert.Equal(t, tc.ok, ok, tc.name)
		assert.Equal(t, tc.domain, d, tc.name)
	}
}

func TestParse(t *testing.T) {
	type tCase struct {
		name     string
		input    string
		expected Reference
		err      error
	}

	for _, tc := range []tCase{
		{
			name:  "Normal",
			input: "docker.io/library/nginx:alpine",
			expected: taggedReference{
				Named: namedReference{"docker.io/library/nginx"},
				tag:   "alpine",
			},
			err: nil,
		}, {
			name:  "Localhost registry",
			input: "localhost:80/nginx:alpine",
			expected: taggedReference{
				Named: namedReference{"localhost:80/nginx"},
				tag:   "alpine",
			},
			err: nil,
		}, {
			name:     " : in path",
			input:    "localhost:80/nginx:nginx/alpine",
			expected: namedReference{"localhost:80/nginx:nginx/alpine"},
			err:      nil,
		}, {
			name:     "Contains scheme",
			input:    "http://docker.io/library/nginx:alpine",
			expected: nil,
			err:      ErrInvalid,
		}, {
			name:     "Contains query",
			input:    "docker.io/library/nginx?tag=alpine",
			expected: nil,
			err:      ErrInvalid,
		}, {
			name:     "Contains fragment",
			input:    "docker.io/library/nginx#tag=alpine",
			expected: nil,
			err:      ErrInvalid,
		}, {
			name:  "Punycode",
			input: "xn--bcher-kva.tld/redis:3",
			expected: taggedReference{
				Named: namedReference{"xn--bcher-kva.tld/redis"},
				tag:   "3",
			},
			err: nil,
		}, {
			name:  "Digest",
			input: "registry.hub.docker.com/library/busybox@sha256:1669a6aa7350e1cdd28f972ddad5aceba2912f589f19a090ac75b7083da748db",
			expected: digestReference{
				Named:  namedReference{"registry.hub.docker.com/library/busybox"},
				digest: "sha256:1669a6aa7350e1cdd28f972ddad5aceba2912f589f19a090ac75b7083da748db",
			},
			err: nil,
		},
	} {
		ref, err := Parse(tc.input)
		assert.Equal(t, tc.err, err, tc.name)
		assert.Equal(t, tc.expected, ref, tc.name)
	}
}
