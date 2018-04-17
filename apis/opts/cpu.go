package opts

import (
	"fmt"
)

// ValidateCPUPeriod validates CPU options for container.
func ValidateCPUPeriod(period int64) error {
	if period == 0 {
		return nil
	}
	if period < 1000 || period > 1000000 {
		return fmt.Errorf("CPU cfs period  %d cannot be less than 1ms (i.e. 1000) or larger than 1s (i.e. 1000000)", period)
	}
	return nil
}

// ValidateCPUQuota validates CPU options for container.
func ValidateCPUQuota(quota int64) error {
	if quota == 0 {
		return nil
	}
	if quota < 1000 {
		return fmt.Errorf("CPU cfs quota %d cannot be less than 1ms (i.e. 1000)", quota)
	}
	return nil
}
