package v1alpha2

import (
	"fmt"
	"os"
	"strings"

	"github.com/alibaba/pouch/cri/config"

	"github.com/cri-o/ocicni/pkg/ocicni"
	"github.com/sirupsen/logrus"
	runtime "k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2"
)

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
				logrus.Errorf("failed to detroy network for sandbox %q: %v", podNetwork.ID, err)
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

// toCNIPortMappings converts CRI port mappings to CNI.
func toCNIPortMappings(criPortMappings []*runtime.PortMapping) []ocicni.PortMapping {
	var portMappings []ocicni.PortMapping
	for _, mapping := range criPortMappings {
		if mapping.HostPort <= 0 {
			continue
		}
		portMappings = append(portMappings, ocicni.PortMapping{
			HostPort:      mapping.HostPort,
			ContainerPort: mapping.ContainerPort,
			Protocol:      strings.ToLower(mapping.Protocol.String()),
			HostIP:        mapping.HostIp,
		})
	}
	return portMappings
}

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
