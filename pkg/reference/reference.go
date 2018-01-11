package reference

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalid is used to return error if reference is invalid
	ErrInvalid = errors.New("invalid reference")

	// defaultTag is latest if there is no tag
	defaultTag = "latest"
)

// Ref represents image name.
type Ref struct {
	Name string
	Tag  string
}

// String returns reference in string.
func (r Ref) String() string {
	// abandon tag if Name is numeric ID
	if isImageID(r.Name) && r.Tag == defaultTag {
		return r.Name
	}

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
	return Ref{Name: s, Tag: tag}, nil
}
