package ocicni

import "github.com/cri-o/ocicni/pkg/ocicni"

// CniMgr as an interface defines all operations against CNI.
type CniMgr interface {
	// Name returns the plugin's name. This will be used when searching
	// for a plugin by name, e.g.
	Name() string

	// SetUpPodNetwork is the method called after the sandbox container of the
	// pod has been created but before the other containers of the pod
	// are launched.
	SetUpPodNetwork(podNetwork *ocicni.PodNetwork) error

	// TearDownPodNetwork is the method called before a pod's sandbox container will be deleted.
	TearDownPodNetwork(podNetwork *ocicni.PodNetwork) error

	// GetPodNetworkStatus is the method called to obtain the ipv4 or ipv6 addresses of the pod sandbox.
	GetPodNetworkStatus(netnsPath string) (string, error)

	// Status returns error if the network plugin is in error state.
	Status() error
}
