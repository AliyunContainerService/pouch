package reference

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	type tCase struct {
		name     string
		input    string
		expected Ref
		err      error
	}

	for _, tc := range []tCase{
		{
			name:  "Normal",
			input: "docker.io/library/nginx:alpine",
			expected: Ref{
				Name: "docker.io/library/nginx",
				Tag:  "alpine",
			},
			err: nil,
		}, {
			name:  "Localhost registry",
			input: "localhost:80/nginx:alpine",
			expected: Ref{
				Name: "localhost:80/nginx",
				Tag:  "alpine",
			},
			err: nil,
		}, {
			name:  " : in path",
			input: "localhost:80/nginx:nginx/alpine",
			expected: Ref{
				Name: "localhost:80/nginx:nginx/alpine",
				Tag:  "latest",
			},
			err: nil,
		}, {
			name:     "Contains scheme",
			input:    "http://docker.io/library/nginx:alpine",
			expected: Ref{},
			err:      ErrInvalid,
		}, {
			name:     "Contains query",
			input:    "docker.io/library/nginx?tag=alpine",
			expected: Ref{},
			err:      ErrInvalid,
		}, {
			name:     "Contains fragment",
			input:    "docker.io/library/nginx#tag=alpine",
			expected: Ref{},
			err:      ErrInvalid,
		}, {
			name:  "Punycode",
			input: "xn--bcher-kva.tld/redis:3",
			expected: Ref{
				Name: "xn--bcher-kva.tld/redis",
				Tag:  "3",
			},
			err: nil,
		}, {
			name:  "OnlyRepoName",
			input: "busybox",
			expected: Ref{
				Name: defaultRegistry + "busybox",
				Tag:  "latest",
			},
			err: nil,
		},
	} {
		ref, err := Parse(tc.input)
		assert.Equal(t, tc.err, err, tc.name)
		assert.Equal(t, tc.expected, ref, tc.name)
	}
}
