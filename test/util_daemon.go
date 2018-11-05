package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
	GetRootDir(&DefaultRootDir)
	// DefaultVolumeMountPath defines the default volume mount path.
	DefaultVolumeMountPath = DefaultRootDir + "/volume"
}

// GetRootDir assign the root dir
func GetRootDir(rootdir *string) error {
	resp, err := request.Get("/info")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	got := types.SystemInfo{}
	err = json.NewDecoder(resp.Body).Decode(&got)
	if err != nil {
		return err
	}
	*rootdir = got.PouchRootDir
	return nil
}

// StartDefaultDaemonDebug starts a deamon with default configuration and debug on.
func StartDefaultDaemonDebug(args ...string) (*daemon.Config, error) {
	cfg := daemon.NewConfig()
	cfg.Debug = true

	cfg.NewArgs(args...)

	return &cfg, cfg.StartDaemon()
}

// StartDefaultDaemon starts a deamon with all default configuration and debug off.
func StartDefaultDaemon(args ...string) (*daemon.Config, error) {
	cfg := daemon.NewConfig()
	cfg.Debug = false

	cfg.NewArgs(args...)

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

	// Find the first -l or --listen parameter and use it.
	for _, v := range d.Args {
		if strings.Contains(v, "-l") || strings.Contains(v, "--listen") {
			if strings.Contains(v, "--listen-cri") {
				continue
			}
			if strings.Contains(v, "=") {
				sock = strings.Split(v, "=")[1]
			} else {
				sock = strings.Fields(v)[1]
			}
			break
		}
	}
	args := append(append([]string{"--host"}, sock), cmd...)
	return command.PouchRun(args...)
}

// CreateConfigFile create configuration file and marshal cfg.
func CreateConfigFile(path string, cfg interface{}) error {
	idx := strings.LastIndex(path, "/")
	if _, err := os.Stat(path[0:idx]); os.IsNotExist(err) {
		os.Mkdir(path[0:idx], os.ModePerm)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	s, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	fmt.Fprintf(file, "%s", s)
	file.Sync()

	defer file.Close()
	return nil
}
