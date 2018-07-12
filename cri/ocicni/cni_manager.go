package ocicni

import (
	"fmt"
	"os"

	"github.com/alibaba/pouch/cri/config"

	"github.com/cri-o/ocicni/pkg/ocicni"
	"github.com/sirupsen/logrus"
)

// CniManager is an implementation of interface CniMgr.
type CniManager struct {
	// plugin is used to setup and teardown network when run/stop pod sandbox.
	plugin ocicni.CNIPlugin
}

// NewCniManager initializes a brand new cni manager.
// If initialize failed, return NoopCniManager, we should not make pouchd creashed
// because of the failure of cni manager.
func NewCniManager(cfg *config.Config) CniMgr {
	networkPluginBinDir := cfg.NetworkPluginBinDir
	networkPluginConfDir := cfg.NetworkPluginConfDir

	// Create CNI configuration directory if it doesn't exist to avoid breaking.
	_, err := os.Stat(networkPluginConfDir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(networkPluginConfDir, 0666)
		if err != nil {
			logrus.Errorf("failed to create configuration directory for CNI: %v", err)
			return &NoopCniManager{}
		}
	}

	plugin, err := ocicni.InitCNI(networkPluginConfDir, networkPluginBinDir)
	if err != nil {
		logrus.Errorf("failed to initialize cni manager: %v", err)
		return &NoopCniManager{}
	}

	return &CniManager{
		plugin: plugin,
	}
}

// Name returns the plugin's name. This will be used when searching
// for a plugin by name, e.g.
func (c *CniManager) Name() string {
	return c.plugin.Name()
}

// SetUpPodNetwork is the method called after the sandbox container of the
// pod has been created but before the other containers of the pod
// are launched.
func (c *CniManager) SetUpPodNetwork(podNetwork *ocicni.PodNetwork) error {
	_, err := c.plugin.SetUpPod(*podNetwork)

	defer func() {
		if err != nil {
			// Teardown network if an error returned.
			err := c.plugin.TearDownPod(*podNetwork)
			if err != nil {
				logrus.Errorf("failed to destroy network for sandbox %q: %v", podNetwork.ID, err)
			}
		}
	}()

	if err != nil {
		return fmt.Errorf("failed to setup network for sandbox %q: %v", podNetwork.ID, err)
	}

	return nil
}

// TearDownPodNetwork is the method called before a pod's sandbox container will be deleted.
func (c *CniManager) TearDownPodNetwork(podNetwork *ocicni.PodNetwork) error {
	err := c.plugin.TearDownPod(*podNetwork)
	if err != nil {
		return fmt.Errorf("failed to destroy network for sandbox %q: %v", podNetwork.ID, err)
	}
	return nil
}

// GetPodNetworkStatus is the method called to obtain the ipv4 or ipv6 addresses of the pod sandbox.
func (c *CniManager) GetPodNetworkStatus(netnsPath string) (string, error) {
	// TODO: we need more validation tests.
	podNetwork := ocicni.PodNetwork{
		NetNS: netnsPath,
	}

	ip, err := c.plugin.GetPodNetworkStatus(podNetwork)
	if err != nil {
		return "", fmt.Errorf("failed to get pod network status: %v", err)
	}

	return ip, nil
}

// Status returns error if the network plugin is in error state.
func (c *CniManager) Status() error {
	return c.plugin.Status()
}
