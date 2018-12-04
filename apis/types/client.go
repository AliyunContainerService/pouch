package types

import "github.com/alibaba/pouch/apis/filters"

// ContainerListOptions holds parameters to list containers with.
type ContainerListOptions struct {
	All    bool
	Before string
	Filter filters.Args
	Limit  int64
	Since  string
}
