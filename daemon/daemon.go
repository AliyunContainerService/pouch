package daemon

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"reflect"

	"github.com/alibaba/pouch/apis/server"
	criservice "github.com/alibaba/pouch/cri"
	"github.com/alibaba/pouch/cri/stream"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/ctrd/supervisord"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/events"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/hookplugins"
	"github.com/alibaba/pouch/internal"
	"github.com/alibaba/pouch/network/mode"
	"github.com/alibaba/pouch/pkg/meta"
	"github.com/alibaba/pouch/pkg/system"

	systemddaemon "github.com/coreos/go-systemd/daemon"
	systemdutil "github.com/coreos/go-systemd/util"
	"github.com/sirupsen/logrus"
)

// Daemon refers to a daemon.
type Daemon struct {
	config         *config.Config
	containerStore *meta.Store

	// ctrdDaemon controls containerd process
	ctrdDaemon *supervisord.Daemon

	// ctrdClient is grpc client connecting to the containerd
	ctrdClient      ctrd.APIClient
	containerMgr    mgr.ContainerMgr
	systemMgr       mgr.SystemMgr
	imageMgr        mgr.ImageMgr
	volumeMgr       mgr.VolumeMgr
	networkMgr      mgr.NetworkMgr
	server          server.Server
	containerPlugin hookplugins.ContainerPlugin
	imagePlugin     hookplugins.ImagePlugin
	daemonPlugin    hookplugins.DaemonPlugin
	volumePlugin    hookplugins.VolumePlugin
	criPlugin       hookplugins.CriPlugin
	apiPlugin       hookplugins.APIPlugin
	eventsService   *events.Events
}

// NewDaemon constructs a brand new server.
func NewDaemon(cfg *config.Config) *Daemon {
	containerStore, err := meta.NewStore(meta.Config{
		Driver:  "local",
		BaseDir: path.Join(cfg.HomeDir, "containers"),
		Buckets: []meta.Bucket{
			{
				Name: meta.MetaJSONFile,
				Type: reflect.TypeOf(mgr.Container{}),
			},
		},
	})
	if err != nil {
		logrus.Errorf("failed to create container meta store: %v", err)
		return nil
	}

	// start containerd
	ctrdDaemonOpts := []supervisord.Opt{
		supervisord.WithOOMScore(cfg.OOMScoreAdjust),
		supervisord.WithGRPCAddress(cfg.ContainerdAddr),
	}

	if cfg.ContainerdPath != "" {
		ctrdDaemonOpts = append(ctrdDaemonOpts, supervisord.WithContainerdBinary(cfg.ContainerdPath))
	}

	if cfg.Debug {
		ctrdDaemonOpts = append(ctrdDaemonOpts, supervisord.WithLogLevel("debug"))
		ctrdDaemonOpts = append(ctrdDaemonOpts, supervisord.WithV1RuntimeShimDebug())
	}

	ctrdDaemon, err := supervisord.Start(context.TODO(),
		filepath.Join(cfg.HomeDir, "containerd/root"),
		filepath.Join(cfg.HomeDir, "containerd/state"),
		ctrdDaemonOpts...,
	)
	if err != nil {
		logrus.Errorf("failed to start containerd: %v", err)
		return nil
	}

	// create containerd client
	ctrdClient, err := ctrd.NewClient(
		ctrd.WithRPCAddr(cfg.ContainerdAddr),
		ctrd.WithDefaultNamespace(cfg.DefaultNamespace),
	)
	if err != nil {
		logrus.Errorf("failed to new containerd's client: %v", err)
		return nil
	}

	if cfg.Snapshotter != "" {
		ctrd.SetSnapshotterName(cfg.Snapshotter)
	}

	if err = ctrdClient.CheckSnapshotterValid(ctrd.CurrentSnapshotterName(context.TODO()), cfg.AllowMultiSnapshotter); err != nil {
		logrus.Errorf("failed to check snapshotter driver: %v", err)
		return nil
	}

	logrus.Infof("Snapshotter is set to be %s", ctrd.CurrentSnapshotterName(context.TODO()))

	return &Daemon{
		config:         cfg,
		ctrdClient:     ctrdClient,
		ctrdDaemon:     ctrdDaemon,
		containerStore: containerStore,
	}
}

func (d *Daemon) loadPlugin() error {
	var err error

	// load daemon plugin if exist
	if daemonPlugin := hookplugins.GetDaemonPlugin(); daemonPlugin != nil {
		d.daemonPlugin = daemonPlugin
	}

	// load container plugin if exist
	if containerPlugin := hookplugins.GetContainerPlugin(); containerPlugin != nil {
		d.containerPlugin = containerPlugin
	}

	// load image plugin if exist
	if imagePlugin := hookplugins.GetImagePlugin(); imagePlugin != nil {
		d.imagePlugin = imagePlugin
	}

	// load volume plugin if exist
	if volumePlugin := hookplugins.GetVolumePlugin(); volumePlugin != nil {
		d.volumePlugin = volumePlugin
	}

	// load cri plugin if exist
	if criPlugin := hookplugins.GetCriPlugin(); criPlugin != nil {
		d.criPlugin = criPlugin
	}

	// load api plugin if exist
	if apiPlugin := hookplugins.GetAPIPlugin(); apiPlugin != nil {
		d.apiPlugin = apiPlugin
	}

	if d.daemonPlugin != nil {
		logrus.Infof("invoke pre-start hook in plugin")
		if err = d.daemonPlugin.PreStartHook(); err != nil {
			return err
		}
	}

	return nil
}

// Run starts daemon.
func (d *Daemon) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := d.loadPlugin(); err != nil {
		return err
	}

	// initializes runtimes real path.
	if err := initialRuntime(d.config.HomeDir, d.config.Runtimes); err != nil {
		return err
	}

	d.eventsService = events.NewEvents()

	imageMgr, err := internal.GenImageMgr(d.config, d)
	if err != nil {
		return err
	}
	d.imageMgr = imageMgr

	systemMgr, err := internal.GenSystemMgr(d.config, d)
	if err != nil {
		return err
	}
	d.systemMgr = systemMgr

	volumeMgr, err := internal.GenVolumeMgr(d.config, d)
	if err != nil {
		return err
	}
	d.volumeMgr = volumeMgr

	containerMgr, err := internal.GenContainerMgr(ctx, d)
	if err != nil {
		return err
	}
	d.containerMgr = containerMgr

	// just register containers information here to let
	// networkMgr to use.
	if err := containerMgr.Load(ctx); err != nil {
		return err
	}

	networkMgr, err := internal.GenNetworkMgr(d.config, d)
	if err != nil {
		return err
	}
	d.networkMgr = networkMgr
	containerMgr.(*mgr.ContainerManager).NetworkMgr = networkMgr

	// after initialize network manager, try to recover all
	// running containers
	if err := containerMgr.Restore(ctx); err != nil {
		return err
	}

	if err := d.addSystemLabels(); err != nil {
		return err
	}

	// init base network
	err = d.networkInit(ctx)
	if err != nil {
		return err
	}

	// set image proxy
	ctrd.SetImageProxy(d.config.ImageProxy)

	criStreamRouterCh := make(chan stream.Router)
	criReadyCh := make(chan bool)
	criStopCh := make(chan error)

	go criservice.RunCriService(d.config, d.containerMgr, d.imageMgr, d.volumeMgr, d.criPlugin, criStreamRouterCh, criStopCh, criReadyCh)

	streamRouter := <-criStreamRouterCh

	d.server = server.Server{
		Config:          d.config,
		ContainerMgr:    containerMgr,
		SystemMgr:       systemMgr,
		ImageMgr:        imageMgr,
		VolumeMgr:       volumeMgr,
		NetworkMgr:      networkMgr,
		StreamRouter:    streamRouter,
		ContainerPlugin: d.containerPlugin,
		APIPlugin:       d.apiPlugin,
	}

	httpReadyCh := make(chan bool)
	httpCloseCh := make(chan struct{})
	go func() {
		if err := d.server.Start(httpReadyCh); err != nil {
			logrus.Errorf("failed to start http server: %v", err)
		}
		close(httpCloseCh)
	}()

	httpReady := <-httpReadyCh
	criReady := <-criReadyCh

	if httpReady && criReady {
		notifySystemd()
	}

	// close the ready channel
	close(httpReadyCh)
	close(criReadyCh)

	err = <-criStopCh
	if err != nil {
		return err
	}

	// Stop pouchd if the server stopped
	<-httpCloseCh
	logrus.Infof("HTTP server stopped")

	return nil
}

// Shutdown stops daemon.
func (d *Daemon) Shutdown() error {
	var errMsg string

	if err := d.server.Stop(); err != nil {
		errMsg = fmt.Sprintf("%s\n", err.Error())
	}

	logrus.Debugf("Start cleanup containerd...")
	if err := d.ctrdClient.Cleanup(); err != nil {
		errMsg = fmt.Sprintf("%s\n", err.Error())
	}

	if err := d.ctrdDaemon.Stop(); err != nil {
		errMsg = fmt.Sprintf("%s\n", err.Error())
	}

	if errMsg != "" {
		return fmt.Errorf("failed to shutdown pouchd: %s", errMsg)
	}
	return nil
}

// Config gets config of daemon.
func (d *Daemon) Config() *config.Config {
	return d.config
}

// CtrMgr gets manager of container.
func (d *Daemon) CtrMgr() mgr.ContainerMgr {
	return d.containerMgr
}

// ImgMgr gets manager of image.
func (d *Daemon) ImgMgr() mgr.ImageMgr {
	return d.imageMgr
}

// VolMgr gets manager of volume.
func (d *Daemon) VolMgr() mgr.VolumeMgr {
	return d.volumeMgr
}

// NetMgr gets manager of network.
func (d *Daemon) NetMgr() mgr.NetworkMgr {
	return d.networkMgr
}

// Containerd gets containerd client.
func (d *Daemon) Containerd() ctrd.APIClient {
	return d.ctrdClient
}

// MetaStore gets store of meta.
func (d *Daemon) MetaStore() *meta.Store {
	return d.containerStore
}

func (d *Daemon) networkInit(ctx context.Context) error {
	return mode.NetworkModeInit(ctx, d.config.NetworkConfig, d.networkMgr)
}

// ContainerPlugin returns the container plugin fetched from shared file
func (d *Daemon) ContainerPlugin() hookplugins.ContainerPlugin {
	return d.containerPlugin
}

// ImagePlugin returns the container plugin fetched from shared file
func (d *Daemon) ImagePlugin() hookplugins.ImagePlugin {
	return d.imagePlugin
}

// ShutdownPlugin invoke pre-stop method in daemon plugin if exist
func (d *Daemon) ShutdownPlugin() error {
	if d.daemonPlugin != nil {
		logrus.Infof("invoke pre-stop hook in plugin")
		if err := d.daemonPlugin.PreStopHook(); err != nil {
			logrus.Errorf("stop prehook execute error %v", err)
		}
	}
	return nil
}

// EventsService gets Events instance
func (d *Daemon) EventsService() *events.Events {
	return d.eventsService
}

// addSystemLabels adds some system labels to daemon's config.
// Currently, pouchd add node ip and serial number to pouchd with the format:
// node_ip=192.168.0.1
// SN=xxxxx
func (d *Daemon) addSystemLabels() error {
	d.config.Lock()
	defer d.config.Unlock()
	if d.config.Labels == nil {
		d.config.Labels = make([]string, 0)
	}
	// get node IP
	nodeIP := system.GetNodeIP()
	d.config.Labels = append(d.config.Labels, fmt.Sprintf("node_ip=%s", nodeIP))

	// get serial number
	serialNo := system.GetSerialNumber()
	d.config.Labels = append(d.config.Labels, fmt.Sprintf("SN=%s", serialNo))

	return nil
}

func notifySystemd() {
	if !systemdutil.IsRunningSystemd() {
		return
	}

	sent, err := systemddaemon.SdNotify(false, "READY=1")
	if err != nil {
		logrus.Errorf("failed to notify systemd for readiness: %v", err)
	}

	if !sent {
		logrus.Errorf("forgot to set Type=notify in systemd service file?")
	}
}
