package opts

import (
	"fmt"
	"strings"
)

// ParseEnv parses the env param of container.
func ParseEnv(env []string) (map[string]string, error) {
	results := make(map[string]string)
	for _, e := range env {
		fields := strings.SplitN(e, "=", 2)
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid env %s: env must be in format of key=value", e)
		}
		results[fields[0]] = fields[1]
	}

	return results, nil
}

// ValidateEnv verifies the correct of env
func ValidateEnv(map[string]string) error {
	// TODO

	return nil
}
