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
			quotaMaps[".*"] = parts[0]
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

// ParseQuotaID parses quota id configurations of container.
func ParseQuotaID(id string, quotas []string) (string, error) {
	switch len(quotas) {
	case 0:
		if isSetQuotaID(id) {
			return "", fmt.Errorf("invalid to set quota id(%s) without disk-quota", id)
		}
	case 1:
		if isSetQuotaID(id) {
			return id, nil
		}

		parts := strings.Split(quotas[0], "=")
		if len(parts) == 1 {
			return "-1", nil
		}
	default:
		if isSetQuotaID(id) {
			return "", fmt.Errorf("invalid to set quota id(%s) for multi disk-quota", id)
		}
	}

	return id, nil
}

func isSetQuotaID(id string) bool {
	return id != "" && id != "0"
}
