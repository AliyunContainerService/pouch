package main

import (
	"runtime"
)

// testImage used in most of tests.
const testImage = "registry.hub.docker.com/library/busybox:latest"

// IsLinux checks if the OS of test environment is Linux.
func IsLinux() bool {
	return runtime.GOOS == "linux"
}
