package reference

import (
	"errors"
	"strings"

	digest "github.com/opencontainers/go-digest"
)

var (
	// ErrInvalid is used to return error if reference is invalid
	ErrInvalid = errors.New("invalid reference")

	// defaultTag is latest if there is no tag
	defaultTag = "latest"
)

// Parse parses ref into Reference.
func Parse(ref string) (Reference, error) {
	if _, err := digest.Parse(ref); err == nil {
		return digestReference(ref), nil
	}

	return ParseNamedReference(ref)
}

// ParseNamedReference parses ref into Named reference.
func ParseNamedReference(ref string) (Named, error) {
	if ok := regRef.MatchString(ref); !ok {
		return nil, ErrInvalid
	}

	// if ref contains tag information
	if loc := regTag.FindStringIndex(ref); loc != nil {
		name, tag := ref[:loc[0]], ref[loc[0]+1:]

		return taggedReference{
			Named: namedReference{name},
			tag:   tag,
		}, nil
	}
	return namedReference{ref}, nil
}

// WithDefaultTagIfMissing adds default tag "latest" for the Named reference if
// the named is not Tagged.
func WithDefaultTagIfMissing(named Named) Named {
	if _, ok := named.(Tagged); !ok {
		return taggedReference{
			Named: named,
			tag:   defaultTag,
		}
	}

	return named
}

// Domain retrieves domain information.
func Domain(named string) (string, bool) {
	i := strings.IndexRune(named, '/')

	// FIXME: The domain should contain the . or :, how to handle the case
	// which image name contains . or :?
	if i == -1 || !strings.ContainsAny(named, ".:") {
		return "", false
	}
	return named[:i], true
}
