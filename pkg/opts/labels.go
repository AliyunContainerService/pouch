package opts

import (
	"fmt"
	"strings"
)

// ParseLabels parses the labels params of container.
func ParseLabels(labels []string) (map[string]string, error) {
	results := make(map[string]string)
	for _, label := range labels {
		fields, err := parseLabel(label)
		if err != nil {
			return nil, err
		}
		k, v := fields[0], fields[1]
		results[k] = v
	}
	return results, nil
}

func parseLabel(label string) ([]string, error) {
	fields := strings.SplitN(label, "=", 2)
	if len(fields) != 2 {
		return nil, fmt.Errorf("invalid label %s: label must be in format of key=value", label)
	}
	return fields, nil
}

// ValidateLabels verifies the correct of labels
func ValidateLabels(map[string]string) error {
	// TODO
	return nil
}
