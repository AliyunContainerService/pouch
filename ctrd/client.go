package ctrd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alibaba/pouch/pkg/scheduler"

	"github.com/containerd/containerd"
	"github.com/sirupsen/logrus"
)

const (
	unixSocketPath                = "/run/containerd/containerd.sock"
	defaultGrpcClientPoolCapacity = 5
	defaultMaxStreamsClient       = 100
)

// Config represents the config used to communicated with containerd.
type Config struct {
	Address string
	// GrpcClientPoolCapacity is the capacity of grpc client pool.
	GrpcClientPoolCapacity int
	// MaxStreamsClient records the max number of concurrent streams
	MaxStreamsClient int
}

// Client is the client side the daemon holds to communicate with containerd.
type Client struct {
	mu sync.RWMutex
	Config
	watch *watch
	lock  *containerLock

	// containerd grpc pool
	pool      []scheduler.Factory
	scheduler scheduler.Scheduler

	hooks []func(string, *Message) error
}

// NewClient connect to containerd.
func NewClient(cfg Config) (APIClient, error) {
	if cfg.Address == "" {
		cfg.Address = unixSocketPath
	}

	if cfg.GrpcClientPoolCapacity <= 0 {
		cfg.GrpcClientPoolCapacity = defaultGrpcClientPoolCapacity
	}

	if cfg.MaxStreamsClient <= 0 {
		cfg.MaxStreamsClient = defaultMaxStreamsClient
	}

	client := &Client{
		Config: cfg,
		lock: &containerLock{
			ids: make(map[string]struct{}),
		},
		watch: &watch{
			containers: make(map[string]*containerPack),
		},
	}

	for i := 0; i < cfg.GrpcClientPoolCapacity; i++ {
		cli, err := newWrapperClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create containerd client: %v", err)
		}
		client.pool = append(client.pool, cli)
	}

	logrus.Infof("success to create %d containerd clients, connect to: %s", cfg.GrpcClientPoolCapacity, cfg.Address)

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
