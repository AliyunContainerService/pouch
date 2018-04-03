package opts

import (
	"fmt"
	"strings"
)

// ParseAnnotation parses runtime annotations format.
func ParseAnnotation(annotations []string) (map[string]string, error) {
	specAnnotation := make(map[string]string)

	for _, annotation := range annotations {
		splits := strings.Split(annotation, "=")
		if len(splits) != 2 || splits[0] == "" || splits[1] == "" {
			return nil, fmt.Errorf("invalid format for spec annotation: %s, correct format should be key=value, neither should be nil", annotation)
		}

		specAnnotation[splits[0]] = splits[1]
	}

	return specAnnotation, nil
}

// ValidateAnnotation validate the correctness of spec annotation param of a container.
func ValidateAnnotation(annotations map[string]string) error {
	// TODO

	return nil
}
