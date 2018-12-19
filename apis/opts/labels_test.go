package opts

import (
	"fmt"
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
			input: []string{"a=b", "a=bb"},
			expected: result{
				labels: nil,
				err:    fmt.Errorf("conflicted labels a=bb and a=b"),
			},
		},
		{
			input: []string{"ThisIsALableWithoutEqualMark"},
			expected: result{
				labels: nil,
				err:    fmt.Errorf("invalid label ThisIsALableWithoutEqualMark: label must be in format of key=value"),
			},
		},
	}

	for _, testCase := range testCases {
		labels, err := ParseLabels(testCase.input)
		assert.Equal(t, testCase.expected.err, err)
		assert.Equal(t, testCase.expected.labels, labels)
	}
}
