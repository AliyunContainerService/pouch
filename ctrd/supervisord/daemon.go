package supervisord

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alibaba/pouch/pkg/utils"

	"github.com/BurntSushi/toml"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/defaults"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	cfgFile    = "pouch-containerd.toml"
	binaryName = "containerd"
	pidFile    = "containerd.pid"

	// healthCheckTimeout used to check alive after start up containerd
	healthCheckTimeout = 3 * time.Second

	// stopTimeout used to send SIGTERM to containerd in limit time to shutdown containerd
	// if the containerd is still alive, Stop action will send SIGKILL
	stopTimeout = 15 * time.Second
)

// Opt is used to modify the daemon setting.
type Opt func(*Daemon) error

// Daemon represents the containerd process.
type Daemon struct {
	cfg Config

	pid        int
	binaryName string
	rootDir    string
	stateDir   string
	logger     *logrus.Entry
}

// Address returns containerd grpc address.
func (d *Daemon) Address() string {
	return d.cfg.GRPC.Address
}

// Start starts containerd in background.
func Start(ctx context.Context, rootDir, stateDir string, opts ...Opt) (*Daemon, error) {
	d := &Daemon{
		cfg: Config{
			Root:  rootDir,
			State: stateDir,
			GRPC: GRPCConfig{
				Address: defaults.DefaultAddress,
			},
			Debug: Debug{
				Level:   "info",
				Address: defaults.DefaultDebugAddress,
			},
		},
		pid:        -1,
		binaryName: binaryName,
		rootDir:    rootDir,
		stateDir:   stateDir,
		logger:     logrus.WithField("module", "ctrd-supervisord"),
	}

	for _, opt := range opts {
		if err := opt(d); err != nil {
			return nil, err
		}
	}

	// for pid and address
	if _, err := os.Stat(d.stateDir); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(d.stateDir, 0666); err != nil {
			return nil, fmt.Errorf("failed to setup state dir %s: %v", d.stateDir, err)
		}
	}

	if err := d.runContainerd(); err != nil {
		return nil, err
	}

	if err := d.healthPostCheck(); err != nil {
		return nil, err
	}
	return d, nil
}

// Stop stops the containerd in 15 seconds.
func (d *Daemon) Stop() error {
	if d.pid != -1 {
		syscall.Kill(d.pid, syscall.SIGTERM)

		for i := time.Duration(0); i < stopTimeout; i += time.Second {
			if !utils.IsProcessAlive(d.pid) {
				break
			}
			time.Sleep(time.Second)
		}

		if utils.IsProcessAlive(d.pid) {
			d.logger.WithField("containerd-pid", d.pid).Warnf("containerd didn't stop within 15secs, killing it")
			utils.KillProcess(d.pid)
		}
	}

	os.Remove(d.pidPath())
	os.Remove(d.Address())
	return nil
}

func (d *Daemon) healthPostCheck() error {
	var (
		failureCount  = 0
		maxRetryCount = 3
		client        *containerd.Client
		err           error
	)

	for ; failureCount < maxRetryCount; failureCount++ {
		time.Sleep(time.Duration(failureCount*500) * time.Millisecond)

		if client == nil {
			client, err = containerd.New(d.Address())
			if err != nil {
				d.logger.WithField("healthcheck", "connection").Warnf("failed to connect to containerd: %v", err)
				client = nil
				continue
			}
		}

		// TODO: should we check the key plugins, like overlayfs?
		tctx, cancel := context.WithTimeout(context.TODO(), healthCheckTimeout)
		ok, err := client.IsServing(tctx)
		cancel()
		if err != nil {
			d.logger.WithField("healthcheck", "serving").Warnf("failed to check IsServing interface: %v", err)
			continue
		}

		if !ok {
			d.logger.WithField("healthcheck", "serving").Warnf("containerd is not serving")
			continue
		}

		client.Close()
		d.logger.WithField("containerd-pid", d.pid).Infof("success to start containerd")
		return nil
	}

	if utils.IsProcessAlive(d.pid) {
		d.logger.WithField("pid", d.pid).Warnf("try to shutdown containerd because failed to connect containerd")
		utils.KillProcess(d.pid)
	}
	return fmt.Errorf("health post check failed")
}

func (d *Daemon) runContainerd() error {
	pid, err := d.getContainerdPid()
	if err != nil {
		return err
	}

	// NOTE:
	// 1. the stdout of containerd will be redirected to /dev/null
	// 2. how to make sure the address is the same one?
	if pid != -1 {
		logrus.WithField("module", "ctrd").WithField("containerd-pid", pid).Infof("containerd is still running")
		return nil
	}

	// if socket file exists, delete it.
	if _, err := os.Stat(d.Address()); err == nil {
		os.RemoveAll(d.Address())
	}

	// setup the configuration
	if err := d.setupConfig(); err != nil {
		return err
	}

	args := []string{"--config", d.configPath()}
	if d.cfg.Debug.Level != "" {
		args = append(args, "--log-level", d.cfg.Debug.Level)
	}

	// start containerd and redirect the stdout to the pouch daemon
	cmd := exec.Command(d.binaryName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Pdeathsig: syscall.SIGKILL}
	cmd.Env = nil

	// clear the NOTIFY_SOCKET from the env when starting containerd
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "NOTIFY_SOCKET") {
			cmd.Env = append(cmd.Env, e)
		}
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// reap the containerd process when it has been killed
	go func() {
		cmd.Wait()
	}()

	if err := d.setContainerdPid(cmd.Process.Pid); err != nil {
		utils.KillProcess(d.pid)
		return fmt.Errorf("failed to save the pid into %s: %v", d.pidPath(), err)
	}
	return nil
}

func (d *Daemon) setContainerdPid(pid int) error {
	d.pid = pid
	return ioutil.WriteFile(d.pidPath(), []byte(fmt.Sprintf("%d", d.pid)), 0660)
}

func (d *Daemon) getContainerdPid() (int, error) {
	f, err := os.OpenFile(d.pidPath(), os.O_RDWR, 0600)
	if err != nil {
		if os.IsNotExist(err) {
			return -1, nil
		}
		return -1, err
	}
	defer f.Close()

	buf := make([]byte, 8)
	num, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return -1, err
	}

	if num > 0 {
		pid, err := strconv.ParseUint(string(buf[:num]), 10, 64)
		if err != nil {
			return -1, err
		}

		if utils.IsProcessAlive(int(pid)) {
			return int(pid), nil
		}
	}
	return -1, nil
}

func (d *Daemon) setupConfig() error {
	f, err := os.OpenFile(d.configPath(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	return errors.Wrap(enc.Encode(d.cfg), "failed to setup containerd config")
}

func (d *Daemon) pidPath() string {
	return filepath.Join(d.stateDir, pidFile)
}

func (d *Daemon) configPath() string {
	return filepath.Join(d.stateDir, cfgFile)
}
