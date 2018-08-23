package cri

import (
	"fmt"

	criv1alpha1 "github.com/alibaba/pouch/cri/v1alpha1"
	servicev1alpha1 "github.com/alibaba/pouch/cri/v1alpha1/service"
	criv1alpha2 "github.com/alibaba/pouch/cri/v1alpha2"
	servicev1alpha2 "github.com/alibaba/pouch/cri/v1alpha2/service"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/mgr"

	"github.com/sirupsen/logrus"
)

// RunCriService start cri service if pouchd is specified with --enable-cri.
func RunCriService(daemonconfig *config.Config, containerMgr mgr.ContainerMgr, imageMgr mgr.ImageMgr, volumeMgr mgr.VolumeMgr, stopCh chan error, readyCh chan bool) {
	var err error

	defer func() {
		stopCh <- err
		close(stopCh)
	}()
	if !daemonconfig.IsCriEnabled {
		// the CriService has been disabled, so send Ready
		readyCh <- true
		return
	}
	switch daemonconfig.CriConfig.CriVersion {
	case "v1alpha1":
		err = runv1alpha1(daemonconfig, containerMgr, imageMgr, readyCh)
	case "v1alpha2":
		err = runv1alpha2(daemonconfig, containerMgr, imageMgr, volumeMgr, readyCh)
	default:
		readyCh <- false
		err = fmt.Errorf("invalid CRI version,failed to start CRI service")
	}
	return
}

// Start CRI service with CRI version: v1alpha1
func runv1alpha1(daemonconfig *config.Config, containerMgr mgr.ContainerMgr, imageMgr mgr.ImageMgr, readyCh chan bool) error {
	logrus.Infof("Start CRI service with CRI version: v1alpha1")
	criMgr, err := criv1alpha1.NewCriManager(daemonconfig, containerMgr, imageMgr)
	if err != nil {
		readyCh <- false
		return fmt.Errorf("failed to get CriManager with error: %v", err)
	}

	service, err := servicev1alpha1.NewService(daemonconfig, criMgr)
	if err != nil {
		readyCh <- false
		return fmt.Errorf("failed to start CRI service with error: %v", err)
	}

	errChan := make(chan error, 2)
	go func() {
		errChan <- service.Serve()
		logrus.Infof("CRI GRPC server stopped")
	}()

	go func() {
		errChan <- criMgr.StreamServerStart()
		logrus.Infof("CRI Stream server stopped")
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
func runv1alpha2(daemonconfig *config.Config, containerMgr mgr.ContainerMgr, imageMgr mgr.ImageMgr, volumeMgr mgr.VolumeMgr, readyCh chan bool) error {
	logrus.Infof("Start CRI service with CRI version: v1alpha2")
	criMgr, err := criv1alpha2.NewCriManager(daemonconfig, containerMgr, imageMgr, volumeMgr)
	if err != nil {
		readyCh <- false
		return fmt.Errorf("failed to get CriManager with error: %v", err)
	}

	service, err := servicev1alpha2.NewService(daemonconfig, criMgr)
	if err != nil {
		readyCh <- false
		return fmt.Errorf("failed to start CRI service with error: %v", err)
	}

	errChan := make(chan error, 2)
	go func() {
		errChan <- service.Serve()
		logrus.Infof("CRI GRPC server stopped")
	}()

	go func() {
		errChan <- criMgr.StreamServerStart()
		logrus.Infof("CRI Stream server stopped")
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
