package opts

import "fmt"

// TODO: ParseMemorySwappiness

// ValidateMemorySwappiness verifies the correctness of memory-swappiness.
func ValidateMemorySwappiness(memorySwappiness int64) error {
	if memorySwappiness < 0 || memorySwappiness > 100 {
		return fmt.Errorf("invalid memory swappiness: %d (its range is 0-100)", memorySwappiness)
	}
	return nil
}
