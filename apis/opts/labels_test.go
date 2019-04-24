package opts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLabels(t *testing.T) {
	type result struct {
		labels map[string]string
		err    error
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
				err: nil,
			},
		},
		{
			input: []string{"a=b", "a=b"},
			expected: result{
				labels: map[string]string{
					"a": "b",
				},
				err: nil,
			},
		},
		{
			// FIXME: this case should throw error
			input: []string{"a=b", "a=bb"},
			expected: result{
				labels: map[string]string{
					"a": "bb",
				},
				err: nil,
			},
		},
		// only input key
		{
			input: []string{"ThisIsALableWithoutEqualMark"},
			expected: result{
				labels: map[string]string{
					"ThisIsALableWithoutEqualMark": "",
				},
				err: nil,
			},
		},
		{
			input: []string{},
			expected: result{
				labels: map[string]string{},
				err: nil,
			},
		},
		{
			input: nil,
			expected: result{
				labels: map[string]string{},
				err: nil,
			},
		},
	}

	for _, testCase := range testCases {
		labels, err := ParseLabels(testCase.input)
		assert.Equal(t, testCase.expected.err, err)
		assert.Equal(t, testCase.expected.labels, labels)
	}
}
