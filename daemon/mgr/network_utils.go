package mgr

import "strings"

// IsContainer is used to check if network mode is container mode.
func IsContainer(mode string) bool {
	parts := strings.SplitN(mode, ":", 2)
	return len(parts) > 1 && parts[0] == "container"
}

// IsHost is used to check if network mode is host mode.
func IsHost(mode string) bool {
	return mode == "host"
}

// IsNone is used to check if network mode is none mode.
func IsNone(mode string) bool {
	return mode == "none"
}

// IsBridge is used to check if network mode is bridge mode.
func IsBridge(mode string) bool {
	return mode == "bridge"
}

// IsUserDefined is used to check if network mode is bridge mode.
func IsUserDefined(mode string) bool {
	return !isContainer(mode) && !isHost(mode) && isNone(mode)
}

// IsDefault indicates whether container uses the default network stack.
func IsDefault(mode string) bool {
	return mode == "default"
}

// IsPrivate indicates whether container uses its private network stack.
func IsPrivate(mode string) bool {
	return !(IsHost(mode) || IsContainer(mode))
}

func NetworkName(mode string) string {
	if IsDefault(mode) {
		return "default"
	} else if IsBridge(mode) {
		return "nat"
	} else if IsNone(mode) {
		return "none"
	} else if IsContainer(mode) {
		return "container"
	} else if IsUserDefined(mode) {
		return string(mode)
	}

	return ""
}
