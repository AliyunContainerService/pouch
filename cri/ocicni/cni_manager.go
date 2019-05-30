package ocicni

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"

	"github.com/alibaba/pouch/cri/config"

	cnicurrent "github.com/containernetworking/cni/pkg/types/current"
	"github.com/cri-o/ocicni/pkg/ocicni"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// CniManager is an implementation of interface CniMgr.
type CniManager struct {
	sync.RWMutex
	// plugin is used to setup and teardown network when run/stop pod sandbox.
	plugin ocicni.CNIPlugin
	// runtimeConfigFile is a file to make the runtime config persistent.
	runtimeConfigFile string
	// defaultRuntimeConfig is configuration specific to the default pod network interface.
	defaultRuntimeConfig ocicni.RuntimeConfig
}

// NewCniManager initializes a brand new cni manager.
func NewCniManager(cfg *config.Config) (CniMgr, error) {
	networkPluginBinDir := cfg.NetworkPluginBinDir
	networkPluginConfDir := cfg.NetworkPluginConfDir

	// Create CNI configuration directory if it doesn't exist to avoid breaking.
	if err := os.MkdirAll(networkPluginConfDir, 0755); err != nil {
		return nil, err
	}

	// load runtime config
	runtimeConfig := ocicni.RuntimeConfig{}

	data, err := ioutil.ReadFile(cfg.RuntimeConfigFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to open runtime config file %s: %v", cfg.RuntimeConfigFile, err)
	}

	if len(data) > 0 {
		err = json.Unmarshal(data, &runtimeConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to read from runtime config file %s: %v", cfg.RuntimeConfigFile, err)
		}
	}

	// If defaultNetName is empty, a network name will be automatically selected be used as the default CNI network.
	plugin, err := ocicni.InitCNI("", networkPluginConfDir, networkPluginBinDir)
	if err != nil {
		return nil, err
	}

	return &CniManager{
		plugin:               plugin,
		defaultRuntimeConfig: runtimeConfig,
		runtimeConfigFile:    cfg.RuntimeConfigFile,
	}, nil
}

// Name returns the plugin's name. This will be used when searching
// for a plugin by name, e.g.
func (c *CniManager) Name() string {
	return c.plugin.Name()
}

// GetDefaultNetworkName returns the name of the plugin's default network.
func (c *CniManager) GetDefaultNetworkName() string {
	return c.plugin.GetDefaultNetworkName()
}

// SetUpPodNetwork is the method called after the sandbox container of the
// pod has been created but before the other containers of the pod
// are launched.
func (c *CniManager) SetUpPodNetwork(podNetwork *ocicni.PodNetwork) error {
	c.RLock()
	c.updateDefaultRuntimeConfig(podNetwork)
	c.RUnlock()

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

// updateDefaultRuntimeConfig set some config of the pod default network interface.
// only set podCIDR now.
func (c *CniManager) updateDefaultRuntimeConfig(podNetwork *ocicni.PodNetwork) {
	if len(c.defaultRuntimeConfig.IpRanges) == 0 {
		return
	}

	if podNetwork.RuntimeConfig == nil {
		podNetwork.RuntimeConfig = make(map[string]ocicni.RuntimeConfig)
	}

	defaultNetworkName := c.GetDefaultNetworkName()

	if _, exist := podNetwork.RuntimeConfig[defaultNetworkName]; !exist {
		podNetwork.RuntimeConfig[defaultNetworkName] = ocicni.RuntimeConfig{}
	}

	if len(podNetwork.RuntimeConfig[defaultNetworkName].IpRanges) == 0 {
		conf := podNetwork.RuntimeConfig[defaultNetworkName]
		conf.IpRanges = c.defaultRuntimeConfig.IpRanges
		podNetwork.RuntimeConfig[defaultNetworkName] = conf
	}
}

// TearDownPodNetwork is the method called before a pod's sandbox container will be deleted.
func (c *CniManager) TearDownPodNetwork(podNetwork *ocicni.PodNetwork) error {
	c.RLock()
	c.updateDefaultRuntimeConfig(podNetwork)
	c.RUnlock()

	// perform the teardown network operation whatever to
	// give CNI Plugin a chance to perform some operations
	err := c.plugin.TearDownPod(*podNetwork)
	if err == nil {
		return nil
	}

	// if netNSPath is not found, dont return error.
	if _, err = os.Stat(podNetwork.NetNS); err != nil {
		if os.IsNotExist(err) {
			logrus.Warnf("failed to find network namespace file %s of sandbox %s", podNetwork.NetNS, podNetwork.ID)
			return nil
		}
		return err
	}
	return errors.Wrapf(err, "failed to destroy network for sandbox %q", podNetwork.ID)
}

// GetPodNetworkStatus is the method called to obtain the ipv4 or ipv6 addresses of the pod sandbox.
func (c *CniManager) GetPodNetworkStatus(netnsPath string) (string, error) {
	// TODO: we need more validation tests.
	podNetwork := ocicni.PodNetwork{
		NetNS: netnsPath,
	}

	var err error
	results, err := c.plugin.GetPodNetworkStatus(podNetwork)
	if err != nil {
		return "", fmt.Errorf("failed to get pod network status: %v", err)
	}

	if len(results) == 0 {
		return "", fmt.Errorf("failed to get pod network status for nil result")
	}

	result, ok := results[0].(*cnicurrent.Result)
	if !ok {
		return "", fmt.Errorf("failed to get pod network status for wrong result: %+v", results[0])
	}

	if len(result.IPs) == 0 {
		return "", fmt.Errorf("failed to get pod network status for nil IP")
	}

	ip := result.IPs[0].Address.IP.String()
	return ip, nil
}

// Status returns error if the network plugin is in error state.
func (c *CniManager) Status() error {
	return c.plugin.Status()
}

const (
	// CNIChangeEventPodCIDR is a subject type to update the podCIDR for the default pod network.
	CNIChangeEventPodCIDR = "pod-cidr-change"
)

// Event handle the changes of CNI.
// only support updatePodCIDR now.
func (c *CniManager) Event(subject string, detail interface{}) error {
	c.Lock()
	defer c.Unlock()

	var err error
	switch subject {
	case CNIChangeEventPodCIDR:
		err = c.updatePodCIDR(detail)
	default:
		err = fmt.Errorf("unknown event subject: %s", subject)
	}

	if err != nil {
		return err
	}

	// save the runtime config
	data, err := json.Marshal(c.defaultRuntimeConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal runtime config: %v", err)
	}

	if err := ioutil.WriteFile(c.runtimeConfigFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config to file %s: %v", c.runtimeConfigFile, err)
	}

	return nil
}

func (c *CniManager) updatePodCIDR(val interface{}) error {
	podCIDR, ok := val.(string)
	if !ok {
		return fmt.Errorf("podCIDR expected string type")
	}

	if _, _, err := net.ParseCIDR(podCIDR); err != nil {
		return fmt.Errorf("\"%s\" is not a valid CIDR value", podCIDR)
	}

	c.defaultRuntimeConfig.IpRanges = [][]ocicni.IpRange{{{Subnet: podCIDR}}}

	return nil
}
