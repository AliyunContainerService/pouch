package opts

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCPUPeriod(t *testing.T) {
	type TestCase struct {
		input    int64
		expected error
	}
	testCases := []TestCase{
		{
			input:    0,
			expected: nil,
		},
		{
			input:    999,
			expected: fmt.Errorf("CPU cfs period  %d cannot be less than 1ms (i.e. 1000) or larger than 1s (i.e. 1000000)", 999),
		},
		{
			input:    1000001,
			expected: fmt.Errorf("CPU cfs period  %d cannot be less than 1ms (i.e. 1000) or larger than 1s (i.e. 1000000)", 1000001),
		},
		{
			input:    -5,
			expected: fmt.Errorf("CPU cfs period  %d cannot be less than 1ms (i.e. 1000) or larger than 1s (i.e. 1000000)", -5),
		},
		{
			input:    1000,
			expected: nil,
		},
		{
			input:    100001,
			expected: nil,
		},
		{
			input:    100000,
			expected: nil,
		},
	}

	for _, testCase := range testCases {
		err := ValidateCPUPeriod(testCase.input)
		assert.Equal(t, testCase.expected, err)
	}
}

func TestValidateCPUQuota(t *testing.T) {
	type TestCase struct {
		quota    int64
		expected error
	}

	testCases := []TestCase{
		{
			quota:    0,
			expected: nil,
		},
		{
			quota:    -1,
			expected: fmt.Errorf("CPU cfs quota %d cannot be less than 1ms (i.e. 1000)", -1),
		},
		{
			quota:    1001,
			expected: nil,
		},
	}

	for _, testCase := range testCases {
		err := ValidateCPUQuota(testCase.quota)
		assert.Equal(t, testCase.expected, err)
	}
}
