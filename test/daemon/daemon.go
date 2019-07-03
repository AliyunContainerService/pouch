package daemon

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/alibaba/pouch/test/command"
	"github.com/alibaba/pouch/test/util"

	"github.com/gotestyourself/gotestyourself/icmd"
)

// For pouch daemon test, we launched another pouch daemon.
const (
	DaemonLog        = "/tmp/pouchd.log"
	PouchdBin        = "pouchd"
	HomeDir          = "/tmp/test/pouch"
	Listen           = "unix:///tmp/test/pouch/pouchd.sock"
	StreamServerPort = "10020"
	ContainerdAdd    = "/tmp/test/pouch/containerd.sock"
	Pidfile          = "/tmp/test/pouch/pouch.pid"
	ConfigFile       = "/tmp/test-config.json"
)

// Config is the configuration of pouch daemon.
type Config struct {
	LogPath string
	LogFile *os.File

	// Daemon startup arguments.
	Args []string

	// Daemon startup config.
	Cfg map[string]interface{}

	// pouchd binary location
	Bin string

	// The following args are all MUST required,
	// in case the new daemon conflicts with existing ones.
	Listen           string
	StreamServerPort string
	HomeDir          string
	ContainerdAddr   string
	Pidfile          string

	// pid of pouchd
	Pid int

	// timeout for starting daemon
	timeout int64

	// if Debug=true, dump daemon log when daemon failed to start
	Debug bool
}

// NewConfig initialize the DConfig with default value.
func NewConfig() Config {
	result := Config{}

	result.Bin = PouchdBin
	result.LogPath = DaemonLog

	result.Args = make([]string, 0, 1)

	result.Listen = Listen
	result.StreamServerPort = StreamServerPort
	result.HomeDir = HomeDir
	result.ContainerdAddr = ContainerdAdd
	result.Pidfile = Pidfile

	result.timeout = 15
	result.Debug = true

	return result
}

// IsDaemonUp checks if the pouchd is launched.
func (d *Config) IsDaemonUp() bool {
	var sock string
	if v, ok := d.Cfg["listen"].([]string); ok {
		for _, host := range v {
			if strings.HasPrefix(host, "unix") {
				sock = host
			}
		}
	}

	if len(sock) != 0 {
		return command.PouchRun("--host", sock, "version").ExitCode == 0
	}

	return command.PouchRun("version").ExitCode == 0
}

// StartDaemon starts pouchd
func (d *Config) StartDaemon() error {
	d.Args = append(d.Args, "--config-file="+ConfigFile)
	cmd := exec.Command(d.Bin, d.Args...)

	// set default config
	if d.Cfg == nil {
		d.Cfg = make(map[string]interface{})
	}
	if _, ok := d.Cfg["listen"]; !ok {
		d.Cfg["listen"] = []string{d.Listen}
	}
	if _, ok := d.Cfg["home-dir"]; !ok {
		d.Cfg["home-dir"] = d.HomeDir
	}
	if _, ok := d.Cfg["containerd"]; !ok {
		d.Cfg["containerd"] = d.ContainerdAddr
	}
	if _, ok := d.Cfg["pidfile"]; !ok {
		d.Cfg["pidfile"] = d.Pidfile
	}
	if _, ok := d.Cfg["cri-config"]; !ok {
		d.Cfg["cri-config"] = map[string]string{
			"stream-server-port": d.StreamServerPort,
		}
	}

	var err error
	if err := CreateConfigFile(ConfigFile, d.Cfg); err != nil {
		return err
	}

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

	if !util.WaitTimeout(time.Duration(d.timeout)*time.Second, d.IsDaemonUp) {
		if d.Debug {
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
	if !d.IsDaemonUp() {
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
}

// CreateConfigFile create configuration file and marshal cfg.
// Merge the default config to it if the default config exists.
func CreateConfigFile(path string, cfg interface{}) error {
	defaultCfg := make(map[string]interface{})
	defaultCfgPath := "/etc/pouch/config.json"
	if _, err := os.Stat(defaultCfgPath); !os.IsNotExist(err) {
		byt, err := ioutil.ReadFile(defaultCfgPath)
		if err != nil {
			return fmt.Errorf("failed to read default config: %v", err)
		}

		if err := json.Unmarshal(byt, &defaultCfg); err != nil {
			return fmt.Errorf("failed to unmarshal default config: %v", err)
		}
	}

	cfgMap := make(map[string]interface{})
	byt, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	if err := json.Unmarshal(byt, &cfgMap); err != nil {
		return fmt.Errorf("failed to unmarshal config: %v", err)
	}

	if cfgMap == nil {
		cfgMap = make(map[string]interface{})
	}

	// merge
	for k, v := range defaultCfg {
		if _, ok := cfgMap[k]; !ok {
			cfgMap[k] = v
		}
	}

	s, err := json.Marshal(cfgMap)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if err = ioutil.WriteFile(path, s, 0644); err != nil {
		return err
	}

	return nil
}
