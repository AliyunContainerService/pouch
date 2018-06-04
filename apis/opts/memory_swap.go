package opts

import units "github.com/docker/go-units"

// ParseMemorySwap parses the memory-swap param of container.
func ParseMemorySwap(memorySwap string) (int64, error) {
	if memorySwap == "" {
		return 0, nil
	}
	if memorySwap == "-1" {
		return -1, nil
	}
	result, err := units.RAMInBytes(memorySwap)
	if err != nil {
		return 0, err
	}
	return result, nil
}

// ValidateMemorySwap verifies the correctness of memory-swap.
func ValidateMemorySwap(memorySwap int64) error {
	// TODO
	return nil
}
