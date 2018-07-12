package opts

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateOOMScore(t *testing.T) {
	type TestCase struct {
		score    int64
		expected error
	}

	testCases := []TestCase{
		{
			score:    -10001,
			expected: fmt.Errorf("oom-score-adj should be in range [-1000, 1000]"),
		},
		{
			score:    10001,
			expected: fmt.Errorf("oom-score-adj should be in range [-1000, 1000]"),
		},
		{
			score:    0,
			expected: nil,
		},
	}

	for _, testCase := range testCases {
		err := ValidateOOMScore(testCase.score)
		assert.Equal(t, testCase.expected, err)
	}
}
