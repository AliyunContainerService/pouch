package reference

import (
	"errors"
	"strings"
)

var (
	// ErrInvalid is used to return error if reference is invalid
	ErrInvalid = errors.New("invalid reference")

	// defaultTag is latest if there is no tag
	defaultTag = "latest"
)

// Parse parses ref into Reference.
func Parse(ref string) (Reference, error) {
	return ParseNamedReference(ref)
}

// ParseNamedReference parses ref into Named reference.
func ParseNamedReference(ref string) (Named, error) {
	if ok := regRef.MatchString(ref); !ok {
		return nil, ErrInvalid
	}

	// if ref contains digest information
	if loc := regDigest.FindStringIndex(ref); loc != nil {
		name, digest := ref[:loc[0]], ref[loc[0]+1:]

		return digestReference{
			Named:  namedReference{name},
			digest: digest,
		}, nil
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

// Domain retrieves domain information. Domain include registry address and
// repository namespace, like registry.hub.docker.com/library/ubuntu.
func Domain(imageRef string) (string, bool) {
	i := strings.LastIndexByte(imageRef, '/')

	// NOTE: in the following two conditions, imageRef doesn't contain domain:
	// 1. No '/' in imageRef.
	// 2. Apart from the name, the rest of imageRef should contain '.' or ':'.
	if i == -1 || !strings.ContainsAny(imageRef[:i], ".:") {
		return "", false
	}
	return imageRef[:i], true
}

// splitHostname splits HostName and RemoteName for the given reference.
// Since we use user defined default registry, if HostName is null, we will return null.
func splitHostname(ref string) (string, string) {
	i := strings.IndexRune(ref, '/')
	if i == -1 || !strings.ContainsAny(ref[:i], ".:") {
		return "", ref
	}
	return ref[:i], ref[i+1:]
}

// IsNameOnly checks if only image repo name only, like busybox.
func IsNameOnly(ref string) bool {
	h, r := splitHostname(ref)
	if h != "" {
		return false
	}

	if strings.Contains(r, "/") {
		return false
	}

	return true
}
