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

// ValidateEnv verifies the correctness of env
// different from the result of ParseEnv can be empty map, a empty env map is invalid
// before invoke ValidateEnv, make sure the parameter not nil nor empty
func ValidateEnv(env map[string]string) (bool, error) {
	if env == nil {
		return false, fmt.Errorf("invalid env %s: env must not be nil", env)
	}
	if len(env) == 0 {
		return false, fmt.Errorf("invalid env %s: env should not be empty", env)
	}
	// TODO if a special key of env should has a range of value,
	// or the key should in a valid key set, it can be validated here

	return true, nil
}
