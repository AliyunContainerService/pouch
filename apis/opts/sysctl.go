package opts

import (
	"fmt"
	"strings"
)

// ParseSysctls parses the sysctl params of container
func ParseSysctls(sysctls []string) (map[string]string, error) {
	results := make(map[string]string)
	for _, sysctl := range sysctls {
		fields, err := parseSysctl(sysctl)
		if err != nil {
			return nil, err
		}
		k, v := fields[0], fields[1]
		results[k] = v
	}
	return results, nil
}

func parseSysctl(sysctl string) ([]string, error) {
	fields := strings.SplitN(sysctl, "=", 2)
	if len(fields) != 2 {
		return nil, fmt.Errorf("invalid sysctl %s: sysctl must be in format of key=value", sysctl)
	}
	return fields, nil
}
