package cri

import (
	"fmt"

	"github.com/alibaba/pouch/cri/stream"
	criv1alpha1 "github.com/alibaba/pouch/cri/v1alpha1"
	servicev1alpha1 "github.com/alibaba/pouch/cri/v1alpha1/service"
	criv1alpha2 "github.com/alibaba/pouch/cri/v1alpha2"
	servicev1alpha2 "github.com/alibaba/pouch/cri/v1alpha2/service"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/hookplugins"

	"github.com/sirupsen/logrus"
)

// RunCriService start cri service if pouchd is specified with --enable-cri.
func RunCriService(daemonconfig *config.Config, containerMgr mgr.ContainerMgr, imageMgr mgr.ImageMgr, volumeMgr mgr.VolumeMgr, criPlugin hookplugins.CriPlugin, streamRouterCh chan stream.Router, stopCh chan error, readyCh chan bool) {
	var err error

	defer func() {
		stopCh <- err
		close(stopCh)
	}()
	if !daemonconfig.IsCriEnabled {
		// the CriService has been disabled, so send Ready and empty Stream Router
		streamRouterCh <- nil
		readyCh <- true
		return
	}
	switch daemonconfig.CriConfig.CriVersion {
	case "v1alpha1":
		err = runv1alpha1(daemonconfig, containerMgr, imageMgr, streamRouterCh, readyCh)
	case "v1alpha2":
		err = runv1alpha2(daemonconfig, containerMgr, imageMgr, volumeMgr, criPlugin, streamRouterCh, readyCh)
	default:
		streamRouterCh <- nil
		readyCh <- false
		err = fmt.Errorf("failed to start CRI service: invalid CRI version %s, expected to be v1alpha1 or v1alpha2", daemonconfig.CriConfig.CriVersion)
	}
	return
}

// Start CRI service with CRI version: v1alpha1
func runv1alpha1(daemonconfig *config.Config, containerMgr mgr.ContainerMgr, imageMgr mgr.ImageMgr, streamRouterCh chan stream.Router, readyCh chan bool) error {
	logrus.Infof("Start CRI service with CRI version: v1alpha1")
	criMgr, err := criv1alpha1.NewCriManager(daemonconfig, containerMgr, imageMgr)
	if err != nil {
		streamRouterCh <- nil
		readyCh <- false
		return fmt.Errorf("failed to get CriManager with error: %v", err)
	}

	service, err := servicev1alpha1.NewService(daemonconfig, criMgr)
	if err != nil {
		streamRouterCh <- nil
		readyCh <- false
		return fmt.Errorf("failed to start CRI service with error: %v", err)
	}

	errChan := make(chan error, 2)
	// If the cri stream server share the port with pouchd,
	// export the its router. Otherwise launch it.
	if daemonconfig.CriConfig.StreamServerReusePort {
		errChan = make(chan error, 1)
		streamRouterCh <- criMgr.StreamRouter()
	} else {
		go func() {
			errChan <- criMgr.StreamServerStart()
			logrus.Infof("CRI Stream server stopped")
		}()
		streamRouterCh <- nil
	}

	go func() {
		errChan <- service.Serve()
		logrus.Infof("CRI GRPC server stopped")
	}()

	// the criservice has set up, send Ready
	readyCh <- true

	// Check for error
	for i := 0; i < cap(errChan); i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	logrus.Infof("CRI service stopped")
	return nil
}

// Start CRI service with CRI version: v1alpha2
func runv1alpha2(daemonconfig *config.Config, containerMgr mgr.ContainerMgr, imageMgr mgr.ImageMgr, volumeMgr mgr.VolumeMgr, criPlugin hookplugins.CriPlugin, streamRouterCh chan stream.Router, readyCh chan bool) error {
	logrus.Infof("Start CRI service with CRI version: v1alpha2")
	criMgr, err := criv1alpha2.NewCriManager(daemonconfig, containerMgr, imageMgr, volumeMgr, criPlugin)
	if err != nil {
		streamRouterCh <- nil
		readyCh <- false
		return fmt.Errorf("failed to get CriManager with error: %v", err)
	}

	service, err := servicev1alpha2.NewService(daemonconfig, criMgr)
	if err != nil {
		streamRouterCh <- nil
		readyCh <- false
		return fmt.Errorf("failed to start CRI service with error: %v", err)
	}

	errChan := make(chan error, 2)
	// If the cri stream server share the port with pouchd,
	// export the its router. Otherwise launch it.
	if daemonconfig.CriConfig.StreamServerReusePort {
		errChan = make(chan error, 1)
		streamRouterCh <- criMgr.StreamRouter()
	} else {
		go func() {
			errChan <- criMgr.StreamServerStart()
			logrus.Infof("CRI Stream server stopped")
		}()
		streamRouterCh <- nil
	}

	go func() {
		errChan <- service.Serve()
		logrus.Infof("CRI GRPC server stopped")
	}()

	// the criservice has set up, send Ready
	readyCh <- true

	// Check for error
	for i := 0; i < cap(errChan); i++ {
		if err := <-errChan; err != nil {
			return err
		}
	}

	logrus.Infof("CRI service stopped")
	return nil
}
