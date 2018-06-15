package opts

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateMemorySwappiness(t *testing.T) {
	type TestCase struct {
		input    int64
		expected error
	}

	testCases := []TestCase{
		{
			input:    -1,
			expected: fmt.Errorf("invalid memory swappiness: %d (its range is 0-100)", -1),
		},
		{
			input:    0,
			expected: nil,
		},
		{
			input:    100,
			expected: nil,
		},
		{
			input:    38,
			expected: nil,
		},
		{
			input:    -5,
			expected: fmt.Errorf("invalid memory swappiness: %d (its range is 0-100)", -5),
		},
		{
			input:    200,
			expected: fmt.Errorf("invalid memory swappiness: %d (its range is 0-100)", 200),
		},
	}

	for _, testCase := range testCases {
		err := ValidateMemorySwappiness(testCase.input)
		assert.Equal(t, testCase.expected, err)
	}
}
