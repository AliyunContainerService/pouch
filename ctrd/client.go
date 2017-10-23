package ctrd

import (
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

	logrus.Infof("success to create containerd's client, connect to: %s", cfg.Address)

	return &Client{
		Config: cfg,
		client: cli,
		lock: &containerLock{
			ids: make(map[string]struct{}),
		},
		watch: &watch{
			containers: make(map[string]containerPack),
			client:     cli,
		},
	}, nil
}

// Close closes the client.
func (c *Client) Close() error {
	return c.client.Close()
}

// Version returns the version of containerd.
func (c *Client) Version() (string, error) {
	return version.Version, nil
}
