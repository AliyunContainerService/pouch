package v1alpha2

import (
	runtime "github.com/alibaba/pouch/cri/apis/v1alpha2"
)

// SandboxMeta represents the sandbox's meta data.
type SandboxMeta struct {
	// ID is the id of sandbox.
	ID string

	// Config is CRI sandbox config.
	Config *runtime.PodSandboxConfig

	// NetNSPath is the network namespace used by the sandbox.
	NetNSPath string

	// Runtime is the runtime of sandbox
	Runtime string
}

// Key returns sandbox's id.
func (meta *SandboxMeta) Key() string {
	return meta.ID
}
