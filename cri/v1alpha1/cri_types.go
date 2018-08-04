package v1alpha1

import (
	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
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

// ContainerMeta reprensets the cri container's meta data.
type ContainerMeta struct {
	// ID is the id of cri container.
	ID string

	// LogPath is the log path of the cri container.
	LogPath string
}

// Key returns container's id.
func (meta *ContainerMeta) Key() string {
	return meta.ID
}
