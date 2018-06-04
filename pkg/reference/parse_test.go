package reference

import (
	"errors"
	"strings"
	"testing"

	digest "github.com/opencontainers/go-digest"
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

	// name@digest
	named = canonicalDigestedReference{
		Named:  namedReference{"pouch"},
		digest: digest.Digest("sha256:dc5f67a48da730d67bf4bfb8824ea8a51be26711de090d6d5a1ffff2723168a3"),
	}
	named = WithDefaultTagIfMissing(named)
	assert.Equal(t, false, strings.Contains(named.String(), "latest"))

	// name:tag@digest
	named = reference{
		Named:  namedReference{"pouch"},
		tag:    "0.4",
		digest: digest.Digest("sha256:dc5f67a48da730d67bf4bfb8824ea8a51be26711de090d6d5a1ffff2723168a3"),
	}
	named = WithDefaultTagIfMissing(named)
	assert.Equal(t, false, strings.Contains(named.String(), "latest"))
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
			name:  "Canonical digested",
			input: "busybox@sha256:1669a6aa7350e1cdd28f972ddad5aceba2912f589f19a090ac75b7083da748db",
			expected: canonicalDigestedReference{
				Named:  namedReference{"busybox"},
				digest: "sha256:1669a6aa7350e1cdd28f972ddad5aceba2912f589f19a090ac75b7083da748db",
			},
			err: nil,
		}, {
			name:  "Digested",
			input: "busybox:1.25@sha256:1669a6aa7350e1cdd28f972ddad5aceba2912f589f19a090ac75b7083da748db",
			expected: reference{
				Named:  namedReference{"busybox"},
				tag:    "1.25",
				digest: "sha256:1669a6aa7350e1cdd28f972ddad5aceba2912f589f19a090ac75b7083da748db",
			},
			err: nil,
		}, {
			name:     "Invalid digested",
			input:    "busybox@sha256:1669a6aa7350e1cdd28f972ddad5aceba2912f589f19a090ac",
			expected: nil,
			err:      errors.New("invalid checksum digest length"),
		}, {
			name:  "Digest ID",
			input: "sha256:1669a6aa7350e1cdd28f972ddad5aceba2912f589f19a090ac",
			expected: taggedReference{
				Named: namedReference{"sha256"},
				tag:   "1669a6aa7350e1cdd28f972ddad5aceba2912f589f19a090ac",
			},
			err: nil,
		},
	} {
		ref, err := Parse(tc.input)
		assert.Equal(t, tc.err, err, tc.name)
		assert.Equal(t, tc.expected, ref, tc.name)
	}
}
