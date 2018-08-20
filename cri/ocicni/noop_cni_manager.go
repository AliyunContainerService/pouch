package ocicni

import "github.com/cri-o/ocicni/pkg/ocicni"

// NoopCniManager is an implementation of interface CniMgr, but makes no operation.
type NoopCniManager struct {
}

// Name of NoopCniManager return the name of plugin as "none".
func (n *NoopCniManager) Name() string {
	return "noop"
}

// SetUpPodNetwork of NoopCniManager makes no operation.
func (n *NoopCniManager) SetUpPodNetwork(podNetwork *ocicni.PodNetwork) error {
	return nil
}

// TearDownPodNetwork of NoopCniManager makes no operation.
func (n *NoopCniManager) TearDownPodNetwork(podNetwork *ocicni.PodNetwork) error {
	return nil
}

// GetPodNetworkStatus of NoopCniManager makes no operation.
func (n *NoopCniManager) GetPodNetworkStatus(netnsPath string) (string, error) {
	return "", nil
}

// Status of NoopCniManager makes no operation.
func (n *NoopCniManager) Status() error {
	return nil
}
