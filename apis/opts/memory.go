package opts

import units "github.com/docker/go-units"

// ParseMemory parses the memory param of container.
func ParseMemory(memory string) (int64, error) {
	if memory == "" {
		return 0, nil
	}
	result, err := units.RAMInBytes(memory)
	if err != nil {
		return 0, err
	}
	return result, nil
}
