package opts

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateCPUPeriod(t *testing.T) {
	tests := []struct {
		name    string
		period  int64
		wantErr bool
	}{
		{name: "test1", period: 0, wantErr: false},
		{name: "test2", period: 999, wantErr: true},
		{name: "test3", period: 1001, wantErr: false},
		{name: "test4", period: 1000001, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCPUPeriod(tt.period)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCPUPeriod() error = %v", err)
			}
		})
	}
}

func TestValidateCPUQuota(t *testing.T) {
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
