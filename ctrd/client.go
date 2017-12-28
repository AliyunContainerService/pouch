package ctrd

import (
	"context"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/version"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const unixSocketPath = "/run/containerd/containerd.sock"

// Config represents the config used to communicated with containerd.
type Config struct {
	Address string
}

// Client is the client side the daemon holds to communicate with containerd.
type Client struct {
	Config
	client *containerd.Client
	watch  *watch
	lock   *containerLock

	hooks []func(string, *Message) error

	// Lease is a new feature of containerd, We use it to avoid that the images
	// are removed by garbage collection. If no lease is defined, the downloaded images will
	// be removed automatically when the container is removed.
	lease *containerd.Lease
}

// NewClient connect to containerd.
func NewClient(cfg Config) (*Client, error) {
	if cfg.Address == "" {
		cfg.Address = unixSocketPath
	}

	options := []containerd.ClientOpt{
		containerd.WithDefaultNamespace("default"),
		// containerd.WithDialOpts([]grpc.DialOption{
		// 	grpc.WithTimeout(time.Second * 5),
		// 	grpc.WithInsecure(),
		// }),
	}
	cli, err := containerd.New(cfg.Address, options...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect containerd")
	}

	// create a new lease or reuse the existed.
	var lease containerd.Lease

	leases, err := cli.ListLeases(context.TODO())
	if err != nil {
		return nil, err
	}
	if len(leases) != 0 {
		lease = leases[0]
	} else {
		if lease, err = cli.CreateLease(context.TODO()); err != nil {
			return nil, err
		}
	}

	logrus.Infof("success to create containerd's client, connect to: %s", cfg.Address)

	return &Client{
		Config: cfg,
		client: cli,
		lease:  &lease,
		lock: &containerLock{
			ids: make(map[string]struct{}),
		},
		watch: &watch{
			containers: make(map[string]*containerPack),
			client:     cli,
		},
	}, nil
}

// SetStopHooks specified the handlers of container exit.
func (c *Client) SetStopHooks(hooks ...func(string, *Message) error) {
	c.watch.hooks = hooks
}

// SetExitHooks specified the handlers of exec process exit.
func (c *Client) SetExitHooks(hooks ...func(string, *Message) error) {
	c.hooks = hooks
}

// Close closes the client.
func (c *Client) Close() error {
	return c.client.Close()
}

// Version returns the version of containerd.
func (c *Client) Version() (string, error) {
	return version.Version, nil
}
