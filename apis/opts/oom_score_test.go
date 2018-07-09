package opts

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateOOMScore(t *testing.T) {
	// TODO
	type TestCase struct {
		input    int64
		expected error
	}

	testCases := []TestCase{
		{
			input:    -1000,
			expected: nil,
		},
		{
			input:    1000,
			expected: nil,
		},
		{
			input:    0,
			expected: nil,
		},
		{
			input:    -2000,
			expected: nil,
		},
		{
			input:    2000,
			expected: fmt.Errorf("oom-score-adj should be in range [-1000, 1000]"),
		},
		{
			input:    200,
			expected: fmt.Errorf("oom-score-adj should be in range [-1000, 1000]"),
		},
	}

	for _, testCase := range testCases {
		err := ValidateOOMScore(testCase.input)
		assert.Equal(t, testCase.expected, err)
	}
}
