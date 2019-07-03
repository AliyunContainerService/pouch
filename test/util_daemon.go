package main

import (
	"encoding/json"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/daemon"
	"github.com/alibaba/pouch/test/request"

	"github.com/gotestyourself/gotestyourself/icmd"
)

var (
	// DefaultRootDir defines the default root dir for pouchd.
	DefaultRootDir string
	// DefaultVolumeMountPath defines the default volume mount path.
	DefaultVolumeMountPath string
)

func init() {
	DefaultRootDir, _ = GetRootDir()
	// DefaultVolumeMountPath defines the default volume mount path.
	DefaultVolumeMountPath = DefaultRootDir + "/volume"
}

// GetRootDir assign the root dir
func GetRootDir() (string, error) {
	resp, err := request.Get("/info")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	got := types.SystemInfo{}
	err = json.NewDecoder(resp.Body).Decode(&got)
	if err != nil {
		return "", err
	}
	return got.PouchRootDir, nil
}

// StartDefaultDaemon starts a daemon with all default configuration and debug off.
func StartDefaultDaemon(configMap map[string]interface{}, args ...string) (*daemon.Config, error) {
	cfg := daemon.NewConfig()
	cfg.Debug = true

	cfg.Cfg = configMap
	cfg.Args = append(cfg.Args, args...)

	return &cfg, cfg.StartDaemon()
}

// RestartDaemon restart daemon
func RestartDaemon(cfg *daemon.Config) error {
	cfg.KillDaemon()
	return cfg.StartDaemon()
}

// RunWithSpecifiedDaemon run pouch command with --host parameter
func RunWithSpecifiedDaemon(d *daemon.Config, cmd ...string) *icmd.Result {
	var sock string
	if v, ok := d.Cfg["listen"].([]string); ok {
		if len(v) > 0 {
			sock = v[0]
		}
	}

	args := append([]string{"--host", sock}, cmd...)
	return command.PouchRun(args...)
}
