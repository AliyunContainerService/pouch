package opts

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMemorySwap(t *testing.T) {
	type result struct {
		memorySwap int64
		err        error
	}
	type TestCase struct {
		input    string
		expected result
	}

	testCases := []TestCase{
		{
			input: "",
			expected: result{
				memorySwap: 0,
				err:        nil,
			},
		},
		{
			input: "-1",
			expected: result{
				memorySwap: -1,
				err:        nil,
			},
		},
		{
			input: "100m",
			expected: result{
				memorySwap: 104857600,
				err:        nil,
			},
		},
		{
			input: "10asdfg",
			expected: result{
				memorySwap: 0,
				err:        fmt.Errorf("invalid size: '%s'", "10asdfg"),
			},
		},
	}

	for _, testCase := range testCases {
		memorySwap, err := ParseMemorySwap(testCase.input)
		assert.Equal(t, testCase.expected.err, err)
		assert.Equal(t, testCase.expected.memorySwap, memorySwap)
	}
}
