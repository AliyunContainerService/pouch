package annotations

// ContainerType values
const (
	// ContainerTypeSandbox represents a pod sandbox container
	ContainerTypeSandbox = "sandbox"

	// ContainerTypeContainer represents a container running within a pod
	ContainerTypeContainer = "container"

	// ContainerType is the container type (sandbox or container) annotation
	ContainerType = "io.kubernetes.cri-o.ContainerType"

	// SandboxName is the sandbox name annotation
	SandboxName = "io.kubernetes.cri-o.SandboxName"

	// KubernetesRuntime is the runtime
	KubernetesRuntime = "io.kubernetes.runtime"
)
