package opts

import (
	"fmt"
	"strings"
)

// ParseEnvs parses the env slice in container's config.
func ParseEnvs(envs []string) ([]string, error) {
	results := []string{}
	for _, env := range envs {
		result, err := parseEnv(env)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// parseEnv parses single elements of env slice.
func parseEnv(env string) (string, error) {
	env = strings.TrimSpace(env)
	if len(env) == 0 {
		return "", fmt.Errorf("invalid env: env cannot be empty")
	}

	arr := strings.SplitN(env, "=", 2)
	if arr[0] == "" {
		return "", fmt.Errorf("invalid env %s: key of env cannot be empty", env)
	}

	if len(arr) == 1 {
		// no matter it is "KEY=" or just "KEY", both are valid.
		return env, nil
	}

	return fmt.Sprintf("%s=%s", arr[0], arr[1]), nil
}

// ValidSliceEnvsToMap converts slice envs to be map
// assuming that the input are always valid with a char of '='.
func ValidSliceEnvsToMap(envs []string) map[string]string {
	results := make(map[string]string)
	for _, env := range envs {
		arr := strings.SplitN(env, "=", 2)
		results[arr[0]] = arr[1]
	}
	return results
}

// ValidMapEnvsToSlice converts valid map envs to slice.
func ValidMapEnvsToSlice(envs map[string]string) []string {
	results := make([]string, 0)
	for key, value := range envs {
		results = append(results, fmt.Sprintf("%s=%s", key, value))
	}
	return results
}

// ValidateEnv verifies the correct of env
func ValidateEnv(map[string]string) error {
	// TODO

	return nil
}
