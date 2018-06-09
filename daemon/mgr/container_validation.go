package mgr

import (
	"fmt"

	"github.com/alibaba/pouch/apis/types"
)

// verifyContainerSetting is to verify the correctness of hostconfig and config.
func (mgr *ContainerManager) verifyContainerSetting(hostConfig *types.HostConfig, config *types.ContainerConfig) error {
	if config != nil {
		// TODO
	}
	return nil
}

// validateLogConfig is used to verify the correctness of log configuration.
func validateLogConfig(logCfg *types.LogConfig) error {
	if logCfg == nil {
		return nil
	}

	switch logCfg.LogDriver {
	case types.LogConfigLogDriverNone, types.LogConfigLogDriverJSONFile:
		return nil
	default:
		return fmt.Errorf("not support (%v) log driver yet", logCfg.LogDriver)
	}
}
