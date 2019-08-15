package mgr

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/logger"
	"github.com/alibaba/pouch/daemon/logger/jsonfile"
	"github.com/alibaba/pouch/daemon/logger/syslog"
	"github.com/alibaba/pouch/pkg/log"
)

const (
	logRootDirKey = "root-dir"
)

func logOptionsForContainerio(c *Container, info logger.Info) (logger.LogDriver, error) {
	cfg := c.HostConfig.LogConfig
	if cfg == nil || cfg.LogDriver == types.LogConfigLogDriverNone {
		return nil, nil
	}

	switch cfg.LogDriver {
	case types.LogConfigLogDriverJSONFile:
		return jsonfile.Init(info)
	case types.LogConfigLogDriverSyslog:
		return syslog.Init(info)
	default:
		log.With(nil).Warnf("not support (%v) log driver yet", cfg.LogDriver)
		return nil, nil
	}
}

// convContainerToLoggerInfo uses logger.Info to wrap container information.
func (mgr *ContainerManager) convContainerToLoggerInfo(c *Container) (logger.Info, error) {
	logCfg := make(map[string]string)
	if cfg := c.HostConfig.LogConfig; cfg != nil && cfg.LogDriver != types.LogConfigLogDriverNone {
		logCfg = cfg.LogOpts
	}

	rootDir, err := mgr.getLogRootDirFromOpt(c, true)
	if err != nil {
		return logger.Info{}, err
	}

	// TODO(fuwei):
	// 1. add more fields into logger.Info
	// 2. separate the logic about retrieving container root dir from mgr.
	return logger.Info{
		LogConfig:        logCfg,
		ContainerID:      c.ID,
		ContainerName:    c.Name,
		ContainerImageID: c.Image,
		ContainerLabels:  c.Config.Labels,
		ContainerEnvs:    c.Config.Env,
		ContainerRootDir: rootDir,
		DaemonName:       "pouchd",
	}, nil
}

// SetContainerLogPath sets the log path of container.
// LogPath would be as a field in `Inspect` response.
func (mgr *ContainerManager) SetContainerLogPath(c *Container) {
	if c.HostConfig.LogConfig == nil {
		return
	}

	// If the logdriver is json-file, the LogPath should be like
	// /var/lib/pouch/containers/5804ee42e505a5d9f30128848293fcb72d8cbc7517310bd24895e82a618fa454/json.log
	if c.HostConfig.LogConfig.LogDriver == "json-file" {
		rootDir, _ := mgr.getLogRootDirFromOpt(c, false)
		c.LogPath = filepath.Join(rootDir, "json.log")
	}
}

func (mgr *ContainerManager) getLogRootDirFromOpt(c *Container, createLogRoot bool) (string, error) {
	// <pouchd-home-dir>/containers/<cid> as default root dir of container log
	rootDir := mgr.Store.Path(c.ID)

	cfg := c.HostConfig.LogConfig
	if cfg == nil || len(cfg.LogOpts) == 0 {
		return rootDir, nil
	}

	specficRootDir, exist := cfg.LogOpts[logRootDirKey]
	if !exist {
		return rootDir, nil
	}

	if !filepath.IsAbs(specficRootDir) {
		return "", fmt.Errorf("As root dir of container log, %s should be abs path", specficRootDir)
	}

	// set <specficRootDir>/<cid> as root dir of container log.
	rootDir = filepath.Join(specficRootDir, c.ID)

	if createLogRoot {
		err := os.MkdirAll(rootDir, 0644)
		if err != nil {
			return "", fmt.Errorf("Failed to mkdirAll %s: %v", rootDir, err)
		}
	}

	return rootDir, nil
}
