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
func RunCriService(daemonconfig *config.Config, containerMgr mgr.ContainerMgr, imageMgr mgr.ImageMgr, stopCh chan error) {
	var err error

	defer func() {
		stopCh <- err
		close(stopCh)
	}()
	if !daemonconfig.IsCriEnabled {
		return
	}
	switch daemonconfig.CriConfig.CriVersion {
	case "v1alpha1":
		err = runv1alpha1(daemonconfig, containerMgr, imageMgr)
	case "v1alpha2":
		err = runv1alpha2(daemonconfig, containerMgr, imageMgr)
	default:
		err = fmt.Errorf("invalid CRI version,failed to start CRI service")
	}
	return
}

// Start CRI service with CRI version: v1alpha1
func runv1alpha1(daemonconfig *config.Config, containerMgr mgr.ContainerMgr, imageMgr mgr.ImageMgr) error {
	logrus.Infof("Start CRI service with CRI version: v1alpha1")
	criMgr, err := criv1alpha1.NewCriManager(daemonconfig, containerMgr, imageMgr)
	if err != nil {
		return fmt.Errorf("failed to get CriManager with error: %v", err)
	}

	service, err := servicev1alpha1.NewService(daemonconfig, criMgr)
	if err != nil {
		return fmt.Errorf("failed to start CRI service with error: %v", err)
	}

	// TODO: Stop the whole CRI service if any of the critical service exits
	grpcServerCloseCh := make(chan struct{})
	go func() {
		if err := service.Serve(); err != nil {
			logrus.Errorf("failed to start grpc server: %v", err)
		}
		close(grpcServerCloseCh)
	}()

	streamServerCloseCh := make(chan struct{})
	go func() {
		if err := criMgr.StreamServerStart(); err != nil {
			logrus.Errorf("failed to start stream server: %v", err)
		}
		close(streamServerCloseCh)
	}()

	// TODO: refactor it with select
	<-streamServerCloseCh
	logrus.Infof("CRI Stream server stopped")
	<-grpcServerCloseCh
	logrus.Infof("CRI GRPC server stopped")

	logrus.Infof("CRI service stopped")
	return nil
}

// Start CRI service with CRI version: v1alpha2
func runv1alpha2(daemonconfig *config.Config, containerMgr mgr.ContainerMgr, imageMgr mgr.ImageMgr) error {
	logrus.Infof("Start CRI service with CRI version: v1alpha2")
	criMgr, err := criv1alpha2.NewCriManager(daemonconfig, containerMgr, imageMgr)
	if err != nil {
		return fmt.Errorf("failed to get CriManager with error: %v", err)
	}

	service, err := servicev1alpha2.NewService(daemonconfig, criMgr)
	if err != nil {
		return fmt.Errorf("failed to start CRI service with error: %v", err)
	}
	// TODO: Stop the whole CRI service if any of the critical service exits
	grpcServerCloseCh := make(chan struct{})
	go func() {
		if err := service.Serve(); err != nil {
			logrus.Errorf("failed to start grpc server: %v", err)
		}
		close(grpcServerCloseCh)
	}()

	streamServerCloseCh := make(chan struct{})
	go func() {
		if err := criMgr.StreamServerStart(); err != nil {
			logrus.Errorf("failed to start stream server: %v", err)
		}
		close(streamServerCloseCh)
	}()
	// TODO: refactor it with select
	<-streamServerCloseCh
	logrus.Infof("CRI Stream server stopped")
	<-grpcServerCloseCh
	logrus.Infof("CRI GRPC server stopped")

	logrus.Infof("CRI service stopped")
	return nil
}
