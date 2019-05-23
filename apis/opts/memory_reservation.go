package opts

import units "github.com/docker/go-units"

// ParseMemoryReservation parses the memory-reservation param of container.
func ParseMemoryReservation(memoryReservation string) (int64, error) {
	if memoryReservation == "" {
		return 0, nil
	}
	return units.RAMInBytes(memoryReservation)
}
