package opts

import (
	"testing"
	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestValidateCPUPeriod(t *testing.T) {
	// mock data
	var tests = [] struct {
		period int64
		expect error
	}{
		{0, nil},
		{1, fmt.Errorf("CPU cfs period  %d cannot be less than 1ms (i.e. 1000) or larger than 1s (i.e. 1000000)", 1)},
		{999, fmt.Errorf("CPU cfs period  %d cannot be less than 1ms (i.e. 1000) or larger than 1s (i.e. 1000000)", 999)},
		{1000, nil},
		{999999, nil},
		{1000000, nil},
		{1000001, fmt.Errorf("CPU cfs period  %d cannot be less than 1ms (i.e. 1000) or larger than 1s (i.e. 1000000)", 1000001)},
		{2000000, fmt.Errorf("CPU cfs period  %d cannot be less than 1ms (i.e. 1000) or larger than 1s (i.e. 1000000)", 2000000)},
	}
 
	for _, tt := range tests {
		actual := ValidateCPUPeriod(tt.period)
		
		// compare the result
		assert.Equal(t, tt.expect, actual)
	}
}

func TestValidateCPUQuota(t *testing.T) {
	// mock data
	var tests = [] struct {
		quota int64
		expect error
	}{
		{0, nil},
		{999, fmt.Errorf("CPU cfs quota %d cannot be less than 1ms (i.e. 1000)", 999)},
		{1000, nil},
		{999999, nil},
	}
 
	for _, tt := range tests {
		actual := ValidateCPUQuota(tt.quota)
		
		// compare the result
		assert.Equal(t, tt.expect, actual)
	}
}
