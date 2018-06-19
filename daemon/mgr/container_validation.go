package mgr

import (
	"fmt"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/logger/syslog"
)

// verifyContainerSetting is to verify the correctness of hostconfig and config.
func (mgr *ContainerManager) verifyContainerSetting(hostConfig *types.HostConfig, config *types.ContainerConfig) error {
	if config != nil {
		// TODO
	}
	return nil
}

// validateLogConfig is used to verify the correctness of log configuration.
func (mgr *ContainerManager) validateLogConfig(c *Container) error {
	logCfg := c.HostConfig.LogConfig
	if logCfg == nil {
		return nil
	}

	switch logCfg.LogDriver {
	case types.LogConfigLogDriverNone, types.LogConfigLogDriverJSONFile:
		return nil
	case types.LogConfigLogDriverSyslog:
		info := mgr.convContainerToLoggerInfo(c)
		return syslog.ValidateSyslogOption(info)
	default:
		return fmt.Errorf("not support (%v) log driver yet", logCfg.LogDriver)
	}
}
