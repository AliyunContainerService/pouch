package opts

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMemory(t *testing.T) {
	type result struct {
		memory int64
		err    error
	}
	type TestCase struct {
		input    string
		expected result
	}

	testCases := []TestCase{
		{
			input: "",
			expected: result{
				memory: 0,
				err:    nil,
			},
		},
		{
			input: "0",
			expected: result{
				memory: 0,
				err:    nil,
			},
		},
		{
			input: "100m",
			expected: result{
				memory: 104857600,
				err:    nil,
			},
		},
		{
			input: "10asdfg",
			expected: result{
				memory: 0,
				err:    fmt.Errorf("invalid size: '%s'", "10asdfg"),
			},
		},
	}

	for _, testCase := range testCases {
		memory, err := ParseMemory(testCase.input)
		assert.Equal(t, testCase.expected.err, err)
		assert.Equal(t, testCase.expected.memory, memory)
	}
}
