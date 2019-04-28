package supervisord

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

	// delayRetryTimeout is used to hold for a while if the restart
	// containerd fails
	delayRetryTimeout = 500 * time.Millisecond
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
	waitCh     chan struct{}
	stopCh     chan struct{}
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
		stopCh:     make(chan struct{}),
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
	d.logger.WithField("containerd-pid", d.pid).Infof("success to start containerd")

	go d.monitor()
	return d, nil
}

// Stop stops the containerd in 15 seconds.
func (d *Daemon) Stop() error {
	// stop the monitor
	close(d.stopCh)

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
		return nil
	}

	if ok, err := d.isContainerdProcess(d.pid); err != nil {
		d.logger.Warnf("failed to get containerd process status by pid %v: %v", d.pid, err)
	} else if ok {
		d.logger.WithField("pid", d.pid).Warnf("try to shutdown containerd because failed to connect containerd")
		utils.KillProcess(d.pid)
	}
	return fmt.Errorf("health post check failed")
}

func (d *Daemon) isContainerdProcess(pid int) (bool, error) {
	if !utils.IsProcessAlive(pid) {
		return false, nil
	}

	// get process path by pid, if readlink -f command exit code not equals 0,
	// we can confirm the pid is not owned by containerd.
	output, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
	if err != nil {
		logrus.WithField("module", "ctrd").WithField("containerd-pid", pid).Infof("got err: %v", err)
		return false, nil
	}
	processPath := strings.TrimSpace(string(output))

	// get containerd process path
	output, err = exec.LookPath(d.binaryName)
	if err != nil {
		return false, err
	}
	containerdPath := strings.TrimSpace(string(output))

	return processPath == containerdPath, nil
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
		if err := d.setContainerdPid(pid); err != nil {
			utils.KillProcess(d.pid)
			return fmt.Errorf("failed to save the pid into %s: %v", d.pidPath(), err)
		}
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

	d.waitCh = make(chan struct{})
	// run and wait the containerd process
	pidChan := make(chan int)
	go func() {
		runtime.LockOSThread()
		defer close(d.waitCh)

		if err := cmd.Start(); err != nil {
			logrus.Errorf("containerd failed to start: %v", err)
			pidChan <- -1
			return
		}
		pidChan <- cmd.Process.Pid
		if err := cmd.Wait(); err != nil {
			logrus.Errorf("containerd exits: %v", err)
		}
	}()

	pid = <-pidChan
	if pid == -1 {
		return fmt.Errorf("containerd failed to start")
	}
	if err := d.setContainerdPid(pid); err != nil {
		utils.KillProcess(d.pid)
		return fmt.Errorf("failed to save the pid into %s: %v", d.pidPath(), err)
	}
	return nil
}

// monitor will try to restart containerd if containerd has been killed or
// panic.
//
// NOTE: if retry time is too much and restart still fails, the monitor
// will exit the whole process.
func (d *Daemon) monitor() {
	var (
		maxRetryCount = 10
		count         = 0
	)

	for {
		select {
		case <-d.stopCh:
			d.logger.Info("receiving stop containerd action and stop monitor")
			return
		default:
		}

		if count > maxRetryCount {
			d.logger.Warnf("failed to restart containerd in time and exit whole process")
			os.Exit(1)
		}

		pid, err := d.getContainerdPid()
		if err != nil {
			d.logger.Warnf("failed to get containerd pid and will retry it again: %v", err)
			count++
			continue
		}

		if pid == -1 {
			if d.waitCh != nil {
				select {
				case <-d.waitCh:
					select {
					case <-d.stopCh:
						d.logger.Info("receiving stop containerd action and stop monitor")
						return
					default:
					}
				case <-d.stopCh:
					d.logger.Info("receiving stop containerd action and stop monitor")
					return
				}
			}

			count++
			if err := d.runContainerd(); err != nil {
				d.logger.Warnf("failed to restart containerd and will retry it again: %v", err)
				time.Sleep(delayRetryTimeout)
				continue
			}
		}

		if err := d.healthPostCheck(); err != nil {
			d.logger.Warn("failed to do health check and will retry it again")
			count++
			time.Sleep(delayRetryTimeout)
			continue
		}

		if count != 0 {
			count = 0
			d.logger.WithField("containerd-pid", d.pid).Infof("success to start containerd")
		}
		time.Sleep(delayRetryTimeout)
	}
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

		isAlive, err := d.isContainerdProcess(int(pid))
		if err != nil {
			return -1, err
		} else if !isAlive {
			logrus.WithField("module", "ctrd").WithField("ctrd-supervisord", pid).Infof("previous containerd pid not exist, delete the pid file")
			os.RemoveAll(d.pidPath())
			return -1, nil
		} else {
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
