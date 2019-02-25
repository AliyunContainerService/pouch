package ctrd

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/alibaba/pouch/pkg/scheduler"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/containerd/containerd"
	eventstypes "github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/api/services/introspection/v1"
	"github.com/containerd/containerd/events"
	"github.com/containerd/containerd/plugin"
	"github.com/containerd/containerd/snapshots"
	"github.com/containerd/typeurl"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	unixSocketPath                = "/run/containerd/containerd.sock"
	defaultGrpcClientPoolCapacity = 5
	defaultMaxStreamsClient       = 100
	// PluginStatusOk means plugin status is ok
	PluginStatusOk = "ok"
	// PluginStatusError means plugin status is error
	PluginStatusError = "error"
)

// ErrGetCtrdClient is an error returned when failed to get a containerd grpc client from clients pool.
var ErrGetCtrdClient = errors.New("failed to get a containerd grpc client")

// Client is the client side the daemon holds to communicate with containerd.
type Client struct {
	mu    sync.RWMutex
	watch *watch
	lock  *containerLock

	// containerd grpc pool
	pool      []scheduler.Factory
	scheduler scheduler.Scheduler

	hooks []func(string, *Message) error

	// eventsHooks specified methods that handle containerd events
	eventsHooks []func(context.Context, string, string, map[string]string) error
}

// Plugin is the containerd plugin type
type Plugin struct {
	Type   string
	ID     string
	Status string
}

// NewClient connect to containerd.
func NewClient(opts ...ClientOpt) (APIClient, error) {
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
	}

	for i := 0; i < copts.grpcClientPoolCapacity; i++ {
		cli, err := newWrapperClient(copts.rpcAddr, copts.defaultns, copts.maxStreamsClient)
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

	// start collect containerd events
	go client.collectContainerdEvents()

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
func (c *Client) SetExitHooks(hooks ...func(string, *Message, func() error) error) {
	c.watch.hooks = hooks
}

// SetExecExitHooks specified the handlers of exec process exit.
func (c *Client) SetExecExitHooks(hooks ...func(string, *Message) error) {
	c.hooks = hooks
}

// SetEventsHooks specified the methods to handle the containerd events.
func (c *Client) SetEventsHooks(hooks ...func(context.Context, string, string, map[string]string) error) {
	c.eventsHooks = hooks
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

// Cleanup handle containerd instance exits.
func (c *Client) Cleanup() error {
	// Note(ziren): notify containerd is dead before containerd
	// is really dead
	c.watch.setContainerdDead(true)

	return c.Close()
}

// Plugins return info of containerd plugins
func (c *Client) Plugins(ctx context.Context, filters []string) ([]Plugin, error) {
	cli, err := c.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get a containerd grpc client: %v", err)
	}

	resp, err := cli.client.IntrospectionService().Plugins(ctx, &introspection.PluginsRequest{Filters: filters})
	if err != nil {
		return nil, err
	}

	var (
		plugins = []Plugin{}
	)

	for _, p := range resp.Plugins {
		plugin := Plugin{
			Type:   p.Type,
			ID:     p.ID,
			Status: PluginStatusOk,
		}

		if p.InitErr != nil {
			plugin.Status = PluginStatusError
		}

		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

// collectContainerdEvents collects events generated by containerd.
func (c *Client) collectContainerdEvents() {
	ctx := context.Background()

	// get client
	wrapperCli, err := c.Get(ctx)
	if err != nil {
		logrus.Errorf("failed to get a containerd grpc client: %v", err)
		return
	}
	eventsClient := wrapperCli.client.EventService()

	// set filters for subscribe containerd events,
	// now we only care about task and container events.
	ef := []string{"topic~=task.*", "topic~=container.*"}
	topicsToHandle := []string{TaskOOMEventTopic, TaskExitEventTopic}

	eventCh, errCh := eventsClient.Subscribe(ctx, ef...)

	for {
		// TODO(ziren):need reconnect the event service
		var e *events.Envelope
		select {
		case e = <-eventCh:
		case err := <-errCh:
			if err != nil {
				logrus.Errorf("failed to receive event: %v", err)
			}
			return
		}

		if !utils.StringInSlice(topicsToHandle, e.Topic) || e.Event == nil {
			continue
		}
		var (
			action      string
			containerID string
			attributes  = map[string]string{}
		)

		out, err := typeurl.UnmarshalAny(e.Event)
		if err != nil {
			logrus.Errorf("failed to unmarshal event %s: %v", e.Topic, err)
			continue
		}

		switch e.Topic {
		case TaskExitEventTopic:
			exitEvent, ok := out.(*eventstypes.TaskExit)
			if !ok {
				logrus.Warnf("failed to parse %s event: %#v", TaskExitEventTopic, out)
				continue
			}
			if exitEvent.ID == exitEvent.ContainerID {
				action = "die"
			} else {
				action = "exec_die"
				attributes["execID"] = exitEvent.ID
			}
			containerID = exitEvent.ContainerID
			attributes["exitCode"] = strconv.Itoa(int(exitEvent.ExitStatus))
		case TaskOOMEventTopic:
			oomEvent, ok := out.(*eventstypes.TaskOOM)
			if !ok {
				logrus.Warnf("failed to parse %s event: %#v", TaskOOMEventTopic, out)
				continue
			}

			action = "oom"
			containerID = oomEvent.ContainerID
		default:
			logrus.Debugf("skip event %s: %#v", e.Topic, out)
			continue
		}

		// handles the event
		for _, hook := range c.eventsHooks {
			if err := hook(ctx, containerID, action, attributes); err != nil {
				logrus.Errorf("failed to execute the containerd events hooks: %v", err)
				break
			}
		}
	}
}

// CheckSnapshotterValid checks whether the given snapshotter is valid
func (c *Client) CheckSnapshotterValid(snapshotter string, allowMultiSnapshotter bool) error {
	var (
		driverFound = false
	)

	plugins, err := c.Plugins(context.Background(), []string{fmt.Sprintf("type==%s", plugin.SnapshotPlugin)})
	if err != nil {
		logrus.Errorf("failed to get containerd plugins: %v", err)
		return err
	}

	for _, p := range plugins {
		if p.Status != PluginStatusOk {
			continue
		}

		if p.ID == snapshotter {
			driverFound = true
			continue
		}

		// if allowMultiSnapshotter, ignore check snapshots exist
		if !allowMultiSnapshotter {
			// check if other snapshotter exists snapshots
			exist, err := c.checkSnapshotsExist(p.ID)
			if err != nil {
				return fmt.Errorf("failed to check snapshotter driver %s: %v", p.ID, err)
			}

			if exist {
				return fmt.Errorf("current snapshotter driver is %s, cannot change to %s", p.ID, snapshotter)
			}
		}
	}

	if !driverFound {
		return fmt.Errorf("containerd not support snapshotter driver %s", snapshotter)
	}

	return nil
}

func (c *Client) checkSnapshotsExist(snapshotter string) (existSnapshot bool, err error) {
	fn := func(c context.Context, s snapshots.Info) error {
		existSnapshot = true
		return nil
	}

	err = c.WalkSnapshot(context.Background(), snapshotter, fn)
	return
}
