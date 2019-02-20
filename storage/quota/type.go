package quota

import "regexp"

// RegExp defines the regular expression of disk quota.
type RegExp struct {
	Pattern *regexp.Regexp
	Path    string
	Size    string
	QuotaID uint32
}

// OverlayMount represents the parameters of overlay mount.
type OverlayMount struct {
	Merged string
	Lower  string
	Upper  string
	Work   string
}
