package namespace

import (
	"strings"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Valid validates all linux namespace mode excepts network mode.
func Valid(ns specs.LinuxNamespaceType, mode string) bool {
	parts := strings.SplitN(string(mode), ":", 2)

	switch ns {
	case specs.UserNamespace:
		switch v := parts[0]; v {
		case "", "host":
		default:
			return false
		}

	case specs.PIDNamespace, specs.UTSNamespace:
		switch v := parts[0]; v {
		case "", "host":
		case "container":
			if len(parts) != 2 || parts[1] == "" {
				return false
			}
		default:
			return false
		}

	case specs.IPCNamespace:
		return IsEmpty(mode) || IsNone(mode) || IsPrivate(ns, mode) || IsHost(mode) || IsShareable(mode) || IsContainer(mode)
	}
	return true
}

// IsEmpty indicates whether namespace mode is empty.
func IsEmpty(mode string) bool {
	return mode == ""
}

// IsNone indicates whether container's namespace mode is set to "none".
func IsNone(mode string) bool {
	return mode == "none"
}

// IsHost indicates whether the container shares the host's corresponding namespace.
func IsHost(mode string) bool {
	return mode == "host"
}

// IsShareable indicates whether the containers namespace can be shared with another container.
func IsShareable(mode string) bool {
	return mode == "shareable"
}

// IsContainer indicates whether the container uses another container's corresponding namespace.
func IsContainer(mode string) bool {
	parts := strings.SplitN(mode, ":", 2)
	return len(parts) > 1 && parts[0] == "container"
}

// IsPrivate indicates whether the container uses its own namespace.
func IsPrivate(ns specs.LinuxNamespaceType, mode string) bool {
	switch ns {
	case specs.IPCNamespace:
		return mode == "private"
	case specs.NetworkNamespace, specs.PIDNamespace:
		return !(IsHost(mode) || IsContainer(mode))
	case specs.UserNamespace, specs.UTSNamespace:
		return !(IsHost(mode))
	}
	return false
}

// IsBridge is used to check if network mode is bridge mode.
func IsBridge(mode string) bool {
	return mode == "bridge"
}

// IsUserDefined is used to check if network mode is user-created.
func IsUserDefined(mode string) bool {
	return !IsBridge(mode) && !IsContainer(mode) && !IsHost(mode) && !IsNone(mode)
}

// IsDefault indicates whether container uses the default network stack.
func IsDefault(mode string) bool {
	return mode == "default"
}

// ConnectedContainer is the id or name of the container whose namespace this container share with.
func ConnectedContainer(mode string) string {
	parts := strings.SplitN(mode, ":", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}
