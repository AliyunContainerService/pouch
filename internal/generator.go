package internal

import (
	"context"
	"sync"

	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/meta"
	"github.com/alibaba/pouch/daemon/mgr"
)

// DaemonProvider provides resources which are needed by container manager and are from daemon.
type DaemonProvider interface {
	Config() *config.Config
	Containerd() *ctrd.Client
	ImgMgr() mgr.ImageMgr
	MetaStore() *meta.Store
}

// GenContainerMgr generates a ContainerMgr instance according to config cfg.
func GenContainerMgr(ctx context.Context, d DaemonProvider) (mgr.ContainerMgr, error) {
	var (
		ctrMgr  *mgr.ContainerManager
		ctrOnce sync.Once
		err     error
	)
	ctrOnce.Do(func() {
		ctrMgr, err = mgr.NewContainerManager(ctx, d.MetaStore())
		ctrMgr.Client = d.Containerd()
		ctrMgr.ImageMgr = d.ImgMgr()
	})

	return ctrMgr, err
}

// GenSystemMgr generates a SystemMgr instance according to config cfg.
func GenSystemMgr(cfg config.Config) (mgr.SystemMgr, error) {
	var (
		sysMgr  *mgr.SystemManager
		sysOnce sync.Once
		err     error
	)

	//TODO
	sysOnce.Do(func() {
		sysMgr, err = mgr.NewSystemManager(cfg)
	})

	return sysMgr, err
}

// GenImageMgr generates a SystemMgr instance according to config cfg.
func GenImageMgr(cfg *config.Config, d DaemonProvider) (mgr.ImageMgr, error) {
	var (
		imgMgr  *mgr.ImageManager
		imgOnce sync.Once
		err     error
	)
	imgOnce.Do(func() {
		imgMgr, err = mgr.NewImageManager(cfg, d.Containerd())
	})

	return imgMgr, err
}

// GenVolumeMgr generates a VolumeMgr instance.
func GenVolumeMgr(cfg config.Config, d DaemonProvider) (mgr.VolumeMgr, error) {
	var (
		volMgr  mgr.VolumeMgr
		volOnce sync.Once
		err     error
	)
	volOnce.Do(func() {
		volMgr, err = mgr.NewVolumeManager(d.MetaStore(), cfg.Config)
	})
	return volMgr, err
}
