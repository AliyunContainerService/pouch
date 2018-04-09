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

	Listen  []string
	HomeDir string

	ContainerdAddr string

	ListenCri string

	// pid of pouchd
	Pid int

	// timeout for starting daemon
	timeout int64

	// if Debug=true, dump daemon log when deamon failed to start
	Debug bool
}

// DConfig is the global variable used to pouch daemon test.
var DConfig Config

func init() {
	DConfig.Args = make([]string, 0, 1)
	DConfig.Listen = make([]string, 0, 1)
	DConfig.timeout = 15
	DConfig.Debug = true
}

// NewConfig initialize the DConfig with default value.
func NewConfig() Config {
	result := Config{}

	result.Bin = PouchdBin
	result.LogPath = DaemonLog

	result.Args = make([]string, 0, 1)
	result.Listen = make([]string, 0, 1)

	result.Args = append(result.Args, "--listen="+Listen)
	result.Args = append(result.Args, "--home-dir="+HomeDir)
	result.Args = append(result.Args, "--containerd="+ContainerdAdd)
	result.Args = append(result.Args, "--listen-cri="+ListenCRI)

	result.Listen = append(result.Listen, Listen)

	result.HomeDir = HomeDir
	result.ContainerdAddr = ContainerdAdd
	result.ListenCri = ListenCRI

	result.timeout = 15

	return result
}

// IsDaemonUp checks if the pouchd is launched.
func (d *Config) IsDaemonUp() bool {
	// if pouchd is started with -l option, use the first listen address
	for _, v := range d.Args {
		if strings.Contains(v, "-l") || strings.Contains(v, "--listen") {
			var sock string
			if strings.Contains(v, "=") {
				sock = strings.Split(v, "=")[1]
			} else {
				sock = strings.Fields(v)[1]
			}
			return command.PouchRun("--host", sock, "version").ExitCode == 0
		}
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
			fmt.Printf("Failed to launch pouchd:%v\n", d.Args)
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
		fmt.Printf("failed to read log, err:%s", err)
	}
	fmt.Printf("pouch daemon log contents: %s", content)
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
