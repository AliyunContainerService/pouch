package daemon

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/util"
	"github.com/gotestyourself/gotestyourself/icmd"
)

// For pouch deamon test, we launched another pouch daemon.
const (
	DaemonLog     = "/tmp/pouchd.log"
	PouchdBin     = "pouchd"
	HomeDir       = "/tmp/test/pouch"
	Listen        = "unix:///tmp/test/pouch/pouchd.sock"
	ContainerdAdd = "/tmp/test/pouch/containerd.sock"
	ListenCRI     = "unix:///tmp/test/pouch/pouchcri.sock"
)

// Config is the configuration of pouch daemon.
type Config struct {
	LogPath string
	LogFile *os.File

	// Daemon startup arguments.
	Args []string

	// pouchd binary location
	Bin string

	// The following args are all MUST required,
	// in case the new daemon conflicts with existing ones.
	Listen         string
	HomeDir        string
	ContainerdAddr string
	ListenCri      string

	// pid of pouchd
	Pid int

	// timeout for starting daemon
	timeout int64

	// if Debug=true, dump daemon log when deamon failed to start
	Debug bool
}

// NewConfig initialize the DConfig with default value.
func NewConfig() Config {
	result := Config{}

	result.Bin = PouchdBin
	result.LogPath = DaemonLog

	result.Args = make([]string, 0, 1)

	result.Listen = Listen
	result.HomeDir = HomeDir
	result.ContainerdAddr = ContainerdAdd
	result.ListenCri = ListenCRI

	result.timeout = 15
	result.Debug = true

	return result
}

// NewArgs is used to construct args according to the struct Config and input.
func (d *Config) NewArgs(args ...string) {
	// Append all default configuration to d.Args if they exists
	// For the rest args in parameter, they must follow the pouchd args usage.
	if len(d.Listen) != 0 {
		d.Args = append(d.Args, "--listen="+d.Listen)
	}
	if len(d.HomeDir) != 0 {
		d.Args = append(d.Args, "--home-dir="+d.HomeDir)
	}
	if len(d.ContainerdAddr) != 0 {
		d.Args = append(d.Args, "--containerd="+d.ContainerdAddr)
	}
	if len(d.ListenCri) != 0 {
		d.Args = append(d.Args, "--listen-cri="+d.ListenCri)
	}

	if len(args) != 0 {
		d.Args = append(d.Args, args...)
	}
}

// IsDaemonUp checks if the pouchd is launched.
func (d *Config) IsDaemonUp() bool {
	// if pouchd is started with -l option, use the first listen address
	var sock string

	for _, v := range d.Args {
		if strings.Contains(v, "-l") || strings.Contains(v, "--listen") {
			if strings.Contains(v, "--listen-cri") {
				continue
			}
			if strings.Contains(v, "=") {
				sock = strings.Split(v, "=")[1]
				break
			} else {
				sock = strings.Fields(v)[1]
				break
			}
		}
	}

	for _, v := range d.Args {
		if strings.Contains(v, "--tlsverify") {
			// TODO: need to verify server with TLS
			return true
		}
	}

	if len(sock) != 0 {
		return command.PouchRun("--host", sock, "version").ExitCode == 0
	}

	return command.PouchRun("version").ExitCode == 0
}

// StartDaemon starts pouchd
func (d *Config) StartDaemon() error {
	cmd := exec.Command(d.Bin, d.Args...)

	var err error
	d.LogFile, err = os.Create(d.LogPath)
	if err != nil {
		return fmt.Errorf("failed to create log file %s, err %s", d.LogPath, err)
	}
	// Must not close the outfile
	//defer outfile.Close()

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

	if util.WaitTimeout(time.Duration(d.timeout)*time.Second, d.IsDaemonUp) == false {
		if d.Debug == true {
			d.DumpLog()

			fmt.Printf("\nFailed to launch pouchd:%v\n", d.Args)

			cmd := "ps aux |grep pouchd"
			fmt.Printf("\nList pouchd process:\n%s\n", icmd.RunCommand("sh", "-c", cmd).Combined())

			cmd = "ps aux |grep containerd"
			fmt.Printf("\nList containerd process:\n%s\n", icmd.RunCommand("sh", "-c", cmd).Combined())
		}

		d.KillDaemon()
		return fmt.Errorf("failed to launch pouchd:%v", d.Args)
	}

	return nil
}

// DumpLog prints the daemon log
func (d *Config) DumpLog() {
	d.LogFile.Sync()

	content, err := ioutil.ReadFile(d.LogPath)
	if err != nil {
		fmt.Printf("failed to read log, err: %s\n", err)
	}
	fmt.Printf("pouch daemon log contents:\n %s\n", content)
}

// KillDaemon kill pouchd.
func (d *Config) KillDaemon() {
	if d.IsDaemonUp() == false {
		return
	}

	if d.Pid != 0 {
		// kill pouchd and all other process in its group
		err := syscall.Kill(-d.Pid, syscall.SIGKILL)
		if err != nil {
			fmt.Printf("kill pouchd failed, err:%s", err)
			return
		}

		d.LogFile.Close()
	}
	return
}
