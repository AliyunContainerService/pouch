package opts

import (
	"fmt"
	"strings"
)

// ParseDiskQuota parses diskquota configurations of container.
func ParseDiskQuota(quotas []string) (map[string]string, error) {
	var quotaMaps = make(map[string]string)

	for _, quota := range quotas {
		if quota == "" {
			return nil, fmt.Errorf("invalid format for disk quota: quota cannot be empty string")
		}

		parts := strings.Split(quota, "=")
		switch len(parts) {
		case 1:
			quotaMaps["/"] = parts[0]
		case 2:
			quotaMaps[parts[0]] = parts[1]
		default:
			return nil, fmt.Errorf("invalid format for disk quota: %s", quota)
		}
	}

	return quotaMaps, nil
}

// ValidateDiskQuota verifies diskquota configurations of container.
func ValidateDiskQuota(quotaMaps map[string]string) error {
	// TODO
	return nil
}
