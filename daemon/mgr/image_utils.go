package mgr

import (
	"regexp"
	"strings"

	"github.com/alibaba/pouch/pkg/reference"
)

// addRegistry add default registry if needed.
func (mgr *ImageManager) addRegistry(input string) string {
	// Trim the prefix if the input is image ID with "sha256:".
	// NOTE: we should make it more elegant and comprehensive.
	input = strings.TrimPrefix(input, "sha256:")
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
