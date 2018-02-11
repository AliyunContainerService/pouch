package mgr

import (
	"fmt"
	"os"
	"strings"

	"github.com/alibaba/pouch/cri"

	"github.com/cri-o/ocicni/pkg/ocicni"
	"github.com/sirupsen/logrus"
	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
)

// CniMgr as an interface defines all operations against CNI.
type CniMgr interface {
	// Name returns the plugin's name. This will be used when searching
	// for a plugin by name, e.g.
	Name() string

	// SetUpPodNetwork is the method called after the sandbox container of the
	// pod has been created but before the other containers of the pod
	// are launched.
	SetUpPodNetwork(config *runtime.PodSandboxConfig, id string, netnsPath string) error

	// TearDownPodNetwork is the method called before a pod's sandbox container will be deleted.
	TearDownPodNetwork(config *runtime.PodSandboxConfig) error

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
func NewCniManager(cfg *cri.Config) (*CniManager, error) {
	networkPluginBinDir := cfg.NetworkPluginBinDir
	networkPluginConfDir := cfg.NetworkPluginConfDir

	// Create CNI configuration directory if it doesn't exist to avoid breaking.
	_, err := os.Stat(networkPluginConfDir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(networkPluginConfDir, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to create configuration directory for CNI: %v", err)
		}
	}

	plugin, err := ocicni.InitCNI(networkPluginConfDir, networkPluginBinDir)
	if err != nil {
		return nil, err
	}

	return &CniManager{
		plugin: plugin,
	}, nil
}

// Name returns the plugin's name. This will be used when searching
// for a plugin by name, e.g.
func (c *CniManager) Name() string {
	return c.plugin.Name()
}

// SetUpPodNetwork is the method called after the sandbox container of the
// pod has been created but before the other containers of the pod
// are launched.
func (c *CniManager) SetUpPodNetwork(config *runtime.PodSandboxConfig, id string, netnsPath string) error {
	podNetwork := ocicni.PodNetwork{
		Name:         config.GetMetadata().GetName(),
		Namespace:    config.GetMetadata().GetNamespace(),
		ID:           id,
		NetNS:        netnsPath,
		PortMappings: toCNIPortMappings(config.GetPortMappings()),
	}

	_, err := c.plugin.SetUpPod(podNetwork)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			// Teardown network if an error returned.
			err := c.plugin.TearDownPod(podNetwork)
			if err != nil {
				logrus.Errorf("failed to detroy network for sandbox %q: %v", id, err)
			}
		}
	}()

	return nil
}

// TearDownPodNetwork is the method called before a pod's sandbox container will be deleted.
func (c *CniManager) TearDownPodNetwork(config *runtime.PodSandboxConfig) error {
	return fmt.Errorf("TearDownPodNetwork Not Implemented Yet")
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
	return fmt.Errorf("Status Not Implemented Yet")
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
