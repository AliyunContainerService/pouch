package mgr

import "strings"

// IsContainer is used to check network mode is container mode.
func IsContainer(mode string) bool {
	parts := strings.SplitN(mode, ":", 2)
	return len(parts) > 1 && parts[0] == "container"
}

// IsHost is used to check network mode is host mode.
func IsHost(mode string) bool {
	return mode == "host"
}

// IsNone is used to check network mode is none mode.
func IsNone(mode string) bool {
	return mode == "none"
}

// IsBridge is used to check network mode is bridge mode.
func IsBridge(mode string) bool {
	return mode == "bridge"
}
