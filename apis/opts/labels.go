package opts

import (
	"strings"
)

// ParseLabels parses the labels params of container.
func ParseLabels(labels []string) map[string]string {
	results := make(map[string]string)
	for _, label := range labels {
		fields := parseLabel(label)
		k, v := fields[0], fields[1]
		results[k] = v
	}
	return results
}

func parseLabel(label string) []string {
	fields := strings.SplitN(label, "=", 2)
	// Only input key without value
	if len(fields) == 1 {
		fields = append(fields, "")
	}
	return fields
}
