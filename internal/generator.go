package internal

import (
	"context"

	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/meta"
	"github.com/alibaba/pouch/daemon/mgr"
)

// DaemonProvider provides resources which are needed by container manager and are from daemon.
type DaemonProvider interface {
	Config() *config.Config
	Containerd() *ctrd.Client
	CtrMgr() mgr.ContainerMgr
	ImgMgr() mgr.ImageMgr
	VolMgr() mgr.VolumeMgr
	NetMgr() mgr.NetworkMgr
	MetaStore() *meta.Store
}

// GenContainerMgr generates a ContainerMgr instance according to config cfg.
func GenContainerMgr(ctx context.Context, d DaemonProvider) (mgr.ContainerMgr, error) {
	return mgr.NewContainerManager(ctx, d.MetaStore(), d.Containerd(), d.ImgMgr(), d.VolMgr(), d.NetMgr(), d.Config())
}

// GenSystemMgr generates a SystemMgr instance according to config cfg.
func GenSystemMgr(cfg *config.Config) (mgr.SystemMgr, error) {
	return mgr.NewSystemManager(cfg)
}

// GenImageMgr generates a ImageMgr instance according to config cfg.
func GenImageMgr(cfg *config.Config, d DaemonProvider) (mgr.ImageMgr, error) {
	return mgr.NewImageManager(cfg, d.Containerd())
}

// GenVolumeMgr generates a VolumeMgr instance according to config cfg.
func GenVolumeMgr(cfg *config.Config, d DaemonProvider) (mgr.VolumeMgr, error) {
	return mgr.NewVolumeManager(d.MetaStore(), cfg.VolumeConfig)
}

// GenNetworkMgr generates a NetworkMgr instance according to config cfg.
func GenNetworkMgr(cfg *config.Config, d DaemonProvider) (mgr.NetworkMgr, error) {
	return mgr.NewNetworkManager(cfg, d.MetaStore())
}

// GenCriMgr generates a CriMgr instance.
func GenCriMgr(d DaemonProvider) (mgr.CriMgr, error) {
	return mgr.NewCriManager(d.Config(), d.CtrMgr(), d.ImgMgr())
}
