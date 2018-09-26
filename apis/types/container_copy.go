package types

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
)

// CopyConfig contains request body of Remote API:
// POST "/containers/"+containerID+"/copy"
type CopyConfig struct {
	Resource string
}

// ContainerPathStat is used to encode the header from
// GET "/containers/{name:.*}/archive"
// "Name" is the file or directory name.
type ContainerPathStat struct {
	Name       string          `json:"name"`
	Size       int64           `json:"size"`
	Mode       uint32          `json:"mode"`
	Mtime      strfmt.DateTime `json:"mtime"`
	LinkTarget string          `json:"linkTarget"`
}

// CopyToContainerOptions holds information
// about files to copy into a container
type CopyToContainerOptions struct {
	AllowOverwriteDirWithFile bool
}

// Validate validates ContainerPathStat
func (m *ContainerPathStat) Validate(formats strfmt.Registry) error {
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
