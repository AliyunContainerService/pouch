package plugins

import (
	"errors"
	"fmt"
)

// ErrPluginStatus represents that the plugin status error.
type ErrPluginStatus struct {
	StatusCode int
	Message    string
}

// Error returns the error message.
func (e *ErrPluginStatus) Error() string {
	return fmt.Sprintf("plugin error code: %d, message: %s", e.StatusCode, e.Message)
}

// ErrNotFound represents that the plugin is not found.
var ErrNotFound = errors.New("plugin not found")

// ErrNotImplemented represents that the plugin not implement the given protocol.
var ErrNotImplemented = errors.New("plugin not implement")
