package quota

import "regexp"

// RegExp defines the regular expression of disk quota.
type RegExp struct {
	Pattern *regexp.Regexp
	Path    string
	Size    string
}
