package daemon

import (
	"context"
	"fmt"
	"path"
	"plugin"
	"reflect"

	"github.com/alibaba/pouch/apis/plugins"
	"github.com/alibaba/pouch/apis/server"
	criservice "github.com/alibaba/pouch/cri"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/internal"
	"github.com/alibaba/pouch/network/mode"
	"github.com/alibaba/pouch/pkg/meta"
	"github.com/alibaba/pouch/pkg/system"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Daemon refers to a daemon.
type Daemon struct {
	config          *config.Config
	containerStore  *meta.Store
	containerd      ctrd.APIClient
	containerMgr    mgr.ContainerMgr
	systemMgr       mgr.SystemMgr
	imageMgr        mgr.ImageMgr
	volumeMgr       mgr.VolumeMgr
	networkMgr      mgr.NetworkMgr
	server          server.Server
	containerPlugin plugins.ContainerPlugin
	daemonPlugin    plugins.DaemonPlugin
}

// router represents the router of daemon.
type router struct {
	daemon *Daemon
	*mux.Router
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

	// New containerd client
	containerdBinaryFile := "containerd"
	if cfg.ContainerdPath != "" {
		containerdBinaryFile = cfg.ContainerdPath
	}
	containerd, err := ctrd.NewClient(cfg.HomeDir,
		ctrd.WithDebugLog(cfg.Debug),
		ctrd.WithStartDaemon(true),
		ctrd.WithContainerdBinary(containerdBinaryFile),
		ctrd.WithRPCAddr(cfg.ContainerdAddr),
		ctrd.WithOOMScoreAdjust(cfg.OOMScoreAdjust),
	)
	if err != nil {
		logrus.Errorf("failed to new containerd's client: %v", err)
		return nil
	}

	return &Daemon{
		config:         cfg,
		containerd:     containerd,
		containerStore: containerStore,
	}
}

func loadSymbolByName(p *plugin.Plugin, name string) (plugin.Symbol, error) {
	s, err := p.Lookup(name)
	if err != nil {
		return nil, errors.Wrapf(err, "lookup plugin with name %s error", name)
	}
	return s, nil
}

func (d *Daemon) loadPlugin() error {
	var s plugin.Symbol
	var err error

	if d.config.PluginPath != "" {
		p, err := plugin.Open(d.config.PluginPath)
		if err != nil {
			return errors.Wrapf(err, "load plugin at %s error", d.config.PluginPath)
		}

		//load container plugin if exist
		if s, err = loadSymbolByName(p, "DaemonPlugin"); err != nil {
			return err
		}
		if daemonPlugin, ok := s.(plugins.DaemonPlugin); ok {
			logrus.Infof("setup daemon plugin from %s", d.config.PluginPath)
			d.daemonPlugin = daemonPlugin
		} else if s != nil {
			return fmt.Errorf("not a container plugin at %s %q", d.config.PluginPath, s)
		}

		//load container plugin if exist
		if s, err = loadSymbolByName(p, "ContainerPlugin"); err != nil {
			return err
		}
		if containerPlugin, ok := s.(plugins.ContainerPlugin); ok {
			logrus.Infof("setup container plugin from %s", d.config.PluginPath)
			d.containerPlugin = containerPlugin
		} else if s != nil {
			return fmt.Errorf("not a container plugin at %s %q", d.config.PluginPath, s)
		}
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

	networkMgr, err := internal.GenNetworkMgr(d.config, d)
	if err != nil {
		return err
	}
	d.networkMgr = networkMgr
	containerMgr.(*mgr.ContainerManager).NetworkMgr = networkMgr

	if err := d.addSystemLabels(); err != nil {
		return err
	}

	d.server = server.Server{
		Config:          d.config,
		ContainerMgr:    containerMgr,
		SystemMgr:       systemMgr,
		ImageMgr:        imageMgr,
		VolumeMgr:       volumeMgr,
		NetworkMgr:      networkMgr,
		ContainerPlugin: d.containerPlugin,
	}

	// init base network
	err = d.networkInit(ctx)
	if err != nil {
		return err
	}

	// set image proxy
	ctrd.SetImageProxy(d.config.ImageProxy)

	httpServerCloseCh := make(chan struct{})
	go func() {
		if err := d.server.Start(); err != nil {
			logrus.Errorf("failed to start http server: %v", err)
		}
		close(httpServerCloseCh)
	}()

	criStopCh := make(chan error)
	go criservice.RunCriService(d.config, d.containerMgr, d.imageMgr, criStopCh)

	err = <-criStopCh
	if err != nil {
		return err
	}

	// Stop pouchd if the server stopped
	<-httpServerCloseCh
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

	if err := d.containerd.Cleanup(); err != nil {
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
	return d.containerd
}

// MetaStore gets store of meta.
func (d *Daemon) MetaStore() *meta.Store {
	return d.containerStore
}

func (d *Daemon) networkInit(ctx context.Context) error {
	return mode.NetworkModeInit(ctx, d.config.NetworkConfg, d.networkMgr)
}

// ContainerPlugin returns the container plugin fetched from shared file
func (d *Daemon) ContainerPlugin() plugins.ContainerPlugin {
	return d.containerPlugin
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
