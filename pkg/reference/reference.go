package reference

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	// ErrInvalid is used to return error if reference is invalid
	ErrInvalid = errors.New("invalid reference")

	// defaultTag is latest if there is no tag
	defaultTag = "latest"

	// defaultRegistry is used to add prefix of image if needed.
	defaultRegistry = "registry.hub.docker.com/library/"
)

// Ref represents image name.
type Ref struct {
	Name string
	Tag  string
}

// String returns reference in string.
func (r Ref) String() string {
	return fmt.Sprintf("%s:%s", r.Name, r.Tag)
}

// Parse a string into reference.
func Parse(s string) (Ref, error) {
	if ok := regRef.MatchString(s); !ok {
		return Ref{}, ErrInvalid
	}

	tag := defaultTag
	if loc := regTag.FindStringIndex(s); loc != nil {
		s, tag = s[:loc[0]], s[loc[0]+1:]
	}

	if !hasRegistry(s) {
		return Ref{Name: defaultRegistry + s, Tag: tag}, nil
	}
	return Ref{Name: s, Tag: tag}, nil
}

// FIXME: need refactor this function.
// hasRegistry check whether image name has registry.
func hasRegistry(s string) bool {
	if strings.Contains(s, "/") || isNumericID(s) {
		return true
	}
	return false
}

// isNumericID checks whether input is numeric ID
func isNumericID(input string) bool {
	match, _ := regexp.MatchString("^[0-9a-f]+$", input)
	return match
}
