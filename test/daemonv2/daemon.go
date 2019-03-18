package daemonv2

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"time"

	daemon "github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/util"

	"github.com/gotestyourself/gotestyourself/icmd"
	"github.com/pkg/errors"
)

// New returns the daemon struct.
func New() *Daemon {
	return &Daemon{
		Config: daemon.Config{
			Listen:         []string{Listen},
			HomeDir:        HomeDir,
			ContainerdAddr: ContainerdAdd,
			Pidfile:        Pidfile,
			Debug:          true,
		},
		Bin:        PouchdBin,
		LogPath:    DaemonLog,
		ConfigJSON: ConfigJSON,

		timeout: 15,
	}
}

// IsUp checks the up or not
func (d *Daemon) IsUp() bool {
	return command.PouchRun("--host", d.Config.Listen[0], "version").ExitCode == 0
}

// Clean kills pouch daemon and clean config file
func (d *Daemon) Clean() {
	d.Stop()

	if d.ConfigJSON != "" {
		os.Remove(d.ConfigJSON)
	}

	if d.LogPath != "" {
		os.Remove(d.LogPath)
	}
}

// DumpLog prints the daemon log
func (d *Daemon) DumpLog() {
	d.LogFile.Sync()

	content, err := ioutil.ReadFile(d.LogPath)
	if err != nil {
		fmt.Printf("failed to read log, err: %s\n", err)
	}
	fmt.Printf("pouch daemon log contents:\n %s\n", content)
}

// Restart restarts pouch daemon
func (d *Daemon) Restart() error {
	if err := d.Stop(); err != nil {
		return err
	}

	return d.Start()
}

// Stop stops pouch daemon
func (d *Daemon) Stop() error {
	if !d.IsUp() {
		return nil
	}

	if d.Pid != 0 {
		// kill pouchd and all other process in its group
		err := syscall.Kill(-d.Pid, syscall.SIGKILL)
		if err != nil {
			return errors.Wrap(err, "failed to kill pouchd")
		}
	}

	return d.LogFile.Close()
}

// Start starts pouch daemon with config file, instead of running with arguments
func (d *Daemon) Start() error {
	fd, err := os.OpenFile(d.ConfigJSON, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	if err := json.NewEncoder(fd).Encode(&d.Config); err != nil {
		fd.Close()
		return err
	}
	fd.Close()

	// start pouchd daemon
	cmd := exec.Command(d.Bin, "--config-file", d.ConfigJSON)

	d.LogFile, err = os.OpenFile(d.LogPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create log file %s, err %s", d.LogPath, err)
	}

	mwriter := io.MultiWriter(d.LogFile)
	cmd.Stderr = mwriter
	cmd.Stdout = mwriter

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start cmd %v, err %s", cmd, err)
	}

	// record the pid
	d.Pid = cmd.Process.Pid

	wait := make(chan error)
	go func() {
		wait <- cmd.Wait()
		fmt.Printf("[%d] exiting daemon", d.Pid)
		close(wait)
	}()

	if !util.WaitTimeout(time.Duration(d.timeout)*time.Second, d.IsUp) {
		if d.Config.Debug {
			d.DumpLog()

			fmt.Printf("\nFailed to launch pouchd\n")

			cmd := "ps aux |grep pouchd"
			fmt.Printf("\nList pouchd process:\n%s\n", icmd.RunCommand("sh", "-c", cmd).Combined())

			cmd = "ps aux |grep containerd"
			fmt.Printf("\nList containerd process:\n%s\n", icmd.RunCommand("sh", "-c", cmd).Combined())
		}

		d.Clean()
		return fmt.Errorf("failed to launch pouchd")
	}

	return nil
}

// RunCommand runs pouch command with specified listen
func (d *Daemon) RunCommand(args ...string) *icmd.Result {
	return command.PouchRun(append([]string{"-H", d.Config.Listen[0]}, args...)...)
}
