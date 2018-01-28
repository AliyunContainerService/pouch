package mgr

import (
	"regexp"

	"github.com/alibaba/pouch/pkg/reference"
)

// addRegistry add default registry if needed.
func (mgr *ImageManager) addRegistry(input string) string {
	if isNumericID(input) {
		return input
	}

	if _, ok := reference.Domain(input); ok {
		return input
	}
	return mgr.DefaultRegistry + input
}

// isNumericID checks whether input is numeric ID
func isNumericID(input string) bool {
	match, _ := regexp.MatchString("^[0-9a-f]+$", input)
	return match
}
