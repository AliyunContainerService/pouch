package opts

import units "github.com/docker/go-units"

// ParseShmSize parses size of /dev/shm
func ParseShmSize(size string) (int64, error) {
	if size == "" {
		return 0, nil
	}
	result, err := units.RAMInBytes(size)
	if err != nil {
		return 0, err
	}
	return result, nil
}
