package opts

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateCPUPeriod(t *testing.T) {
	// TODO
}

func TestValidateCPUQuota(t *testing.T) {
	// TODO
	type TestCase struct {
		input    int64
		expected error
	}

	testCases := []TestCase{
		{
			input:    -1,
			expected: fmt.Errorf("CPU cfs quota %d cannot be less than 1ms (i.e. 1000)", -1),
		},
		{
			input:    0,
			expected: nil,
		},
		{
			input:    1000,
			expected: nil,
		},
		{
			input:    500,
			expected: fmt.Errorf("CPU cfs quota %d cannot be less than 1ms (i.e. 1000)", 500),
		},
		{
			input:    -5,
			expected: fmt.Errorf("CPU cfs quota %d cannot be less than 1ms (i.e. 1000)", -5),
		},
		{
			input:    1200,
			expected: nil,
		},
	}

	for _, testCase := range testCases {
		err := ValidateCPUQuota(testCase.input)
		assert.Equal(t, testCase.expected, err)
	}
}
