package mgr

import (
	"regexp"
	"strings"
)

// FIXME: need refactor this function.
// addRegistry add default registry if needed.
func (mgr *ImageManager) addRegistry(input string) string {
	if strings.Contains(input, "/") || isNumericID(input) {
		return input
	}
	return mgr.DefaultRegistry + input
}

// isNumericID checks whether input is numeric ID
func isNumericID(input string) bool {
	match, _ := regexp.MatchString("^[0-9a-f]+$", input)
	return match
}
