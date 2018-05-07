package ctrd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alibaba/pouch/pkg/scheduler"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/containerd/containerd"
	"github.com/sirupsen/logrus"
)

const (
	unixSocketPath                = "/run/containerd/containerd.sock"
	containerdPidFileName         = "containerd.pid"
	defaultGrpcClientPoolCapacity = 5
	defaultMaxStreamsClient       = 100
	containerdShutdownTimeout     = 15 * time.Second
)

// Client is the client side the daemon holds to communicate with containerd.
type Client struct {
	mu    sync.RWMutex
	watch *watch
	lock  *containerLock

	daemonPid      int
	homeDir        string
	rpcAddr        string
	oomScoreAdjust int
	debugLog       bool

	// containerd grpc pool
	pool      []scheduler.Factory
	scheduler scheduler.Scheduler

	hooks []func(string, *Message) error
}

// NewClient connect to containerd.
func NewClient(homeDir string, opts ...ClientOpt) (APIClient, error) {
	// set default value for parameters
	copts := clientOpts{
		rpcAddr:                unixSocketPath,
		grpcClientPoolCapacity: defaultGrpcClientPoolCapacity,
		maxStreamsClient:       defaultMaxStreamsClient,
	}

	for _, opt := range opts {
		if err := opt(&copts); err != nil {
			return nil, err
		}
	}

	client := &Client{
		lock: &containerLock{
			ids: make(map[string]struct{}),
		},
		watch: &watch{
			containers: make(map[string]*containerPack),
		},
		daemonPid:      -1,
		homeDir:        homeDir,
		oomScoreAdjust: copts.oomScoreAdjust,
		debugLog:       copts.debugLog,
		rpcAddr:        copts.rpcAddr,
	}

	// start new containerd instance.
	if copts.startDaemon {
		if err := client.runContainerdDaemon(homeDir, copts); err != nil {
			return nil, err
		}
	}

	for i := 0; i < copts.grpcClientPoolCapacity; i++ {
		cli, err := newWrapperClient(copts.rpcAddr, copts.maxStreamsClient)
		if err != nil {
			return nil, fmt.Errorf("failed to create containerd client: %v", err)
		}
		client.pool = append(client.pool, cli)
	}

	logrus.Infof("success to create %d containerd clients, connect to: %s", copts.grpcClientPoolCapacity, copts.rpcAddr)

	scheduler, err := scheduler.NewLRUScheduler(client.pool)
	if err != nil {
		return nil, fmt.Errorf("failed to create clients pool scheduler")
	}
	client.scheduler = scheduler

	return client, nil
}

// Get will reture an available containerd grpc client,
// Or occurred an error
func (c *Client) Get(ctx context.Context) (*WrapperClient, error) {
	start := time.Now()

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Scheduler returns Factory interface
	factory, err := c.scheduler.Schedule(ctx)
	if err != nil {
		return nil, err
	}

	wrapperCli, ok := factory.(*WrapperClient)
	if !ok {
		return nil, fmt.Errorf("failed to convert Factory interface to *WrapperClient")
	}

	end := time.Now()
	elapsed := end.Sub(start)
	logrus.WithFields(logrus.Fields{
		"elapsed": elapsed,
	}).Debug("Get a grpc client")

	return wrapperCli, nil
}

// SetExitHooks specified the handlers of container exit.
func (c *Client) SetExitHooks(hooks ...func(string, *Message) error) {
	c.watch.hooks = hooks
}

// SetExecExitHooks specified the handlers of exec process exit.
func (c *Client) SetExecExitHooks(hooks ...func(string, *Message) error) {
	c.hooks = hooks
}

// Close closes the client.
func (c *Client) Close() error {
	c.mu.Lock()
	factories := c.pool
	c.pool = nil
	c.mu.Unlock()

	if factories == nil {
		return nil
	}

	var (
		errInfo []string
		err     error
	)

	for _, c := range factories {
		wrapperCli, ok := c.(*WrapperClient)
		if !ok {
			errInfo = append(errInfo, "failed to convert Factory interface to *WrapperClient")
			continue
		}

		if err := wrapperCli.client.Close(); err != nil {
			errInfo = append(errInfo, err.Error())
			continue
		}
	}

	if len(errInfo) > 0 {
		err = fmt.Errorf("failed to close client pool: %s", errInfo)
	}
	return err
}

// Version returns the version of containerd.
func (c *Client) Version(ctx context.Context) (containerd.Version, error) {
	cli, err := c.Get(ctx)
	if err != nil {
		return containerd.Version{}, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	return cli.client.Version(ctx)
}

func (c *Client) runContainerdDaemon(homeDir string, copts clientOpts) error {
	if homeDir == "" {
		return fmt.Errorf("ctrd: containerd home dir should not be empty")
	}

	containerdPath, err := exec.LookPath(copts.containerdBinary)
	if err != nil {
		return fmt.Errorf("failed to find containerd binary %s: %v", copts.containerdBinary, err)
	}

	stateDir := path.Join(homeDir, "containerd/state")
	if _, err := os.Stat(stateDir); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(stateDir, 0666); err != nil {
			return fmt.Errorf("failed to mkdir %s: %v", stateDir, err)
		}
	}

	pidFileName := path.Join(stateDir, containerdPidFileName)
	f, err := os.OpenFile(pidFileName, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 8)
	num, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return err
	}

	if num > 0 {
		pid, err := strconv.ParseUint(string(buf[:num]), 10, 64)
		if err != nil {
			return err
		}
		if utils.IsProcessAlive(int(pid)) {
			logrus.Infof("ctrd: previous instance of containerd still alive (%d)", pid)
			c.daemonPid = int(pid)
			return nil
		}
	}

	// empty container pid file
	_, err = f.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	if err := f.Truncate(0); err != nil {
		return err
	}

	// if socket file exists, delete it.
	if _, err := os.Stat(c.rpcAddr); err == nil {
		os.RemoveAll(c.rpcAddr)
	}

	cmd, err := c.newContainerdCmd(containerdPath)
	if err != nil {
		return err
	}

	if err := utils.SetOOMScore(cmd.Process.Pid, c.oomScoreAdjust); err != nil {
		utils.KillProcess(cmd.Process.Pid)
		return err
	}

	if _, err := f.WriteString(fmt.Sprintf("%d", cmd.Process.Pid)); err != nil {
		utils.KillProcess(cmd.Process.Pid)
		return err
	}

	go cmd.Wait()

	c.daemonPid = cmd.Process.Pid
	return nil
}

func (c *Client) newContainerdCmd(containerdPath string) (*exec.Cmd, error) {
	// Start a new containerd instance
	args := []string{
		"-a", c.rpcAddr,
		"--root", path.Join(c.homeDir, "containerd/root"),
		"--state", path.Join(c.homeDir, "containerd/state"),
		"-l", utils.If(c.debugLog, "debug", "info").(string),
	}

	cmd := exec.Command(containerdPath, args...)
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
		return nil, err
	}

	logrus.Infof("ctrd: new containerd process, pid: %d", cmd.Process.Pid)
	return cmd, nil
}

// Cleanup handle containerd instance exits.
func (c *Client) Cleanup() error {
	if c.daemonPid == -1 {
		return nil
	}

	if err := c.Close(); err != nil {
		return err
	}

	// Ask the daemon to quit
	syscall.Kill(c.daemonPid, syscall.SIGTERM)

	// Wait up to 15secs for it to stop
	for i := time.Duration(0); i < containerdShutdownTimeout; i += time.Second {
		if !utils.IsProcessAlive(c.daemonPid) {
			break
		}
		time.Sleep(time.Second)
	}

	if utils.IsProcessAlive(c.daemonPid) {
		logrus.Warnf("ctrd: containerd (%d) didn't stop within 15secs, killing it\n", c.daemonPid)
		syscall.Kill(c.daemonPid, syscall.SIGKILL)
	}

	// cleanup some files
	os.Remove(path.Join(c.homeDir, "containerd/state", containerdPidFileName))
	os.Remove(c.rpcAddr)

	return nil
}
