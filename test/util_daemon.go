package main

import (
	"strings"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/daemon"

	"github.com/gotestyourself/gotestyourself/icmd"
)

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

// StartDaemonBareWithArgs starts a deamon with all user specified parameter.
func StartDaemonBareWithArgs(cfg *daemon.Config, args ...string) error {
	cfg.NewArgs(args...)

	return cfg.StartDaemon()
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
