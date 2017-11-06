package main

import (
	"runtime"
)

// IsLinux checks if the OS of test environment is Linux.
func IsLinux() bool {
	return runtime.GOOS == "linux"
}
