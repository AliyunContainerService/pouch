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

	// SandboxID is the sandbox id annotation
	SandboxID = "io.kubernetes.cri-o.SandboxID"

	// KubernetesRuntime is the runtime
	KubernetesRuntime = "io.kubernetes.runtime"

	// LxcfsEnabled whether to enable lxcfs for a container
	LxcfsEnabled = "io.kubernetes.lxcfs.enabled"
)
