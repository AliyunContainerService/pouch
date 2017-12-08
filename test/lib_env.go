package main

import (
	"runtime"
)

// testImage used in most of tests.
const testImage = "registry.hub.docker.com/library/busybox:latest"

// default pouch binary
const defaultBinary = "/usr/local/bin/pouch"

// IsLinux checks if the OS of test environment is Linux.
func IsLinux() bool {
	return runtime.GOOS == "linux"
}
