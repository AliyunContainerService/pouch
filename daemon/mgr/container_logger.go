package mgr

import (
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/containerio"
	"github.com/alibaba/pouch/daemon/logger"

	"github.com/sirupsen/logrus"
)

func optionsForContainerio(c *Container) []func(*containerio.Option) {
	optFuncs := make([]func(*containerio.Option), 0, 1)

	cfg := c.HostConfig.LogConfig
	if cfg == nil || cfg.LogDriver == types.LogConfigLogDriverNone {
		return optFuncs
	}

	switch cfg.LogDriver {
	case types.LogConfigLogDriverJSONFile:
		optFuncs = append(optFuncs, containerio.WithJSONFile())
	default:
		logrus.Warnf("not support %v log driver yet", cfg.LogDriver)
	}
	return optFuncs
}

// convContainerToLoggerInfo uses logger.Info to wrap container information.
func (mgr *ContainerManager) convContainerToLoggerInfo(c *Container) logger.Info {
	logCfg := make(map[string]string)
	if cfg := c.HostConfig.LogConfig; cfg != nil && cfg.LogDriver != types.LogConfigLogDriverNone {
		logCfg = cfg.LogOpts
	}

	// TODO:
	// 1. add more fields into logger.Info
	// 2. separate the logic about retrieving container root dir from mgr.
	return logger.Info{
		LogConfig:        logCfg,
		ContainerID:      c.ID,
		ContainerLabels:  c.Config.Labels,
		ContainerRootDir: mgr.Store.Path(c.ID),
	}
}
