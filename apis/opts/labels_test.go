package opts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLabels(t *testing.T) {
	type result struct {
		labels map[string]string
	}

	type TestCase struct {
		input    []string
		expected result
	}

	testCases := []TestCase{
		{
			input: []string{"a=b"},
			expected: result{
				labels: map[string]string{
					"a": "b",
				},
			},
		},
		{
			input: []string{"a=b", "a=b"},
			expected: result{
				labels: map[string]string{
					"a": "b",
				},
			},
		},
		{
			// FIXME: this case should throw error
			input: []string{"a=b", "a=bb"},
			expected: result{
				labels: map[string]string{
					"a": "bb",
				},
			},
		},
		// only input key
		{
			input: []string{"a"},
			expected: result{
				labels: map[string]string{
					"a": "",
				},
			},
		},
		{
			input: []string{"a="},
			expected: result{
				labels: map[string]string{
					"a": "",
				},
			},
		},
		{
			input: []string{"a=b=c"},
			expected: result{
				labels: map[string]string{
					"a": "b=c",
				},
			},
		},
		{
			input: []string{},
			expected: result{
				labels: map[string]string{},
			},
		},
		{
			input: nil,
			expected: result{
				labels: map[string]string{},
			},
		},
	}

	for _, testCase := range testCases {
		labels := ParseLabels(testCase.input)
		assert.Equal(t, testCase.expected.labels, labels)
	}
}
