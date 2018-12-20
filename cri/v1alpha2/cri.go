package v1alpha2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	goruntime "runtime"
	"time"

	"github.com/alibaba/pouch/apis/filters"
	apitypes "github.com/alibaba/pouch/apis/types"
	anno "github.com/alibaba/pouch/cri/annotations"
	runtime "github.com/alibaba/pouch/cri/apis/v1alpha2"
	"github.com/alibaba/pouch/cri/metrics"
	cni "github.com/alibaba/pouch/cri/ocicni"
	"github.com/alibaba/pouch/cri/stream"
	criutils "github.com/alibaba/pouch/cri/utils"
	"github.com/alibaba/pouch/ctrd"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/hookplugins"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/meta"
	"github.com/alibaba/pouch/pkg/reference"
	pkgstreams "github.com/alibaba/pouch/pkg/streams"
	"github.com/alibaba/pouch/pkg/utils"
	util_metrics "github.com/alibaba/pouch/pkg/utils/metrics"
	"github.com/alibaba/pouch/version"

	// NOTE: "golang.org/x/net/context" is compatible with standard "context" in golang1.7+.
	"github.com/cri-o/ocicni/pkg/ocicni"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	pouchRuntimeName         = "pouch"
	kubeletRuntimeAPIVersion = "0.1.0"

	// kubePrefix is used to idenfify the containers/sandboxes on the node managed by kubelet.
	kubePrefix = "k8s"

	// annotationPrefix is used to distinguish between annotations and labels.
	annotationPrefix = "annotation."

	// Internal pouch labels used to identify whether a container is a sandbox
	// or a regular container.
	containerTypeLabelKey       = "io.kubernetes.pouch.type"
	containerTypeLabelSandbox   = "sandbox"
	containerTypeLabelContainer = "container"
	sandboxIDLabelKey           = "io.kubernetes.sandbox.id"
	containerLogPathLabelKey    = "io.kubernetes.container.logpath"

	// sandboxContainerName is a string to include in the pouch container so
	// that users can easily identify the sandboxes.
	sandboxContainerName = "POD"

	// nameDelimiter is used to construct pouch container names.
	nameDelimiter = "_"

	namespaceModeHost = "host"
	namespaceModeNone = "none"

	// resolvConfPath is the abs path of resolv.conf on host or container.
	resolvConfPath = "/etc/resolv.conf"

	// snapshotPlugin implements a snapshotter.
	snapshotPlugin = "io.containerd.snapshotter.v1"

	// networkNotReadyReason is the reason reported when network is not ready.
	networkNotReadyReason = "NetworkPluginNotReady"

	// passthruKey to specify whether a interface is passthru to qemu
	passthruKey = "io.alibaba.pouch.vm.passthru"

	// passthruIP is the IP for container
	passthruIP = "io.alibaba.pouch.vm.passthru.ip"
)

var (
	// Default timeout for stopping container.
	defaultStopTimeout = int64(10)
)

// CriMgr as an interface defines all operations against CRI.
type CriMgr interface {
	// RuntimeServiceServer is interface of CRI runtime service.
	runtime.RuntimeServiceServer

	// ImageServiceServer is interface of CRI image service.
	runtime.ImageServiceServer

	// VolumeServiceServer is interface of CRI volume service.
	runtime.VolumeServiceServer

	// StreamServerStart starts the stream server of CRI.
	StreamServerStart() error

	// StreamStart returns the router of Stream Server.
	StreamRouter() stream.Router
}

// CriManager is an implementation of interface CriMgr.
type CriManager struct {
	ContainerMgr mgr.ContainerMgr
	ImageMgr     mgr.ImageMgr
	VolumeMgr    mgr.VolumeMgr
	CniMgr       cni.CniMgr
	CriPlugin    hookplugins.CriPlugin

	// StreamServer is the stream server of CRI serves container streaming request.
	StreamServer Server

	// SandboxBaseDir is the directory used to store sandbox files like /etc/hosts, /etc/resolv.conf, etc.
	SandboxBaseDir string

	// SandboxImage is the image used by sandbox container.
	SandboxImage string

	// SandboxStore stores the configuration of sandboxes.
	SandboxStore *meta.Store

	// SnapshotStore stores information of all snapshots.
	SnapshotStore *mgr.SnapshotStore

	// imageFSPath is the path to image filesystem.
	imageFSPath string
}

// NewCriManager creates a brand new cri manager.
func NewCriManager(config *config.Config, ctrMgr mgr.ContainerMgr, imgMgr mgr.ImageMgr, volumeMgr mgr.VolumeMgr, criPlugin hookplugins.CriPlugin) (CriMgr, error) {
	var streamServerAddress string
	streamServerPort := config.CriConfig.StreamServerPort
	// If stream server reuse the pouchd's port, extract the ip and port from pouchd's listening addresses.
	if config.CriConfig.StreamServerReusePort {
		streamServerAddress, streamServerPort = extractIPAndPortFromAddresses(config.Listen)
		if streamServerPort == "" {
			return nil, fmt.Errorf("failed to extract stream server's port from pouchd's listening addresses")
		}
	}

	// If the reused pouchd's port is https, the url that stream server return should be with https scheme.
	reuseHTTPSPort := config.CriConfig.StreamServerReusePort && config.TLS.Key != "" && config.TLS.Cert != ""
	streamServer, err := newStreamServer(ctrMgr, streamServerAddress, streamServerPort, reuseHTTPSPort)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream server for cri manager: %v", err)
	}

	c := &CriManager{
		ContainerMgr:   ctrMgr,
		ImageMgr:       imgMgr,
		VolumeMgr:      volumeMgr,
		CriPlugin:      criPlugin,
		StreamServer:   streamServer,
		SandboxBaseDir: path.Join(config.HomeDir, "sandboxes"),
		SandboxImage:   config.CriConfig.SandboxImage,
		SnapshotStore:  mgr.NewSnapshotStore(),
	}
	c.CniMgr, err = cni.NewCniManager(&config.CriConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create cni manager: %v", err)
	}

	c.SandboxStore, err = meta.NewStore(meta.Config{
		Driver:  "local",
		BaseDir: path.Join(config.HomeDir, "sandboxes-meta"),
		Buckets: []meta.Bucket{
			{
				Name: meta.MetaJSONFile,
				Type: reflect.TypeOf(SandboxMeta{}),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox meta store: %v", err)
	}

	c.imageFSPath = imageFSPath(path.Join(config.HomeDir, "containerd/root"), ctrd.CurrentSnapshotterName())
	logrus.Infof("Get image filesystem path %q", c.imageFSPath)

	if !config.CriConfig.DisableCriStatsCollect {
		period := config.CriConfig.CriStatsCollectPeriod
		if period <= 0 {
			return nil, fmt.Errorf("cri stats collect period should > 0")
		}
		snapshotsSyncer := ctrMgr.NewSnapshotsSyncer(
			c.SnapshotStore,
			time.Duration(period)*time.Second,
		)
		snapshotsSyncer.Start()
	} else {
		logrus.Infof("disable cri to collect stats from containerd periodically")
	}

	return NewCriWrapper(c), nil
}

// StreamServerStart starts the stream server of CRI.
func (c *CriManager) StreamServerStart() error {
	return c.StreamServer.Start()
}

// StreamRouter returns the router of Stream Server.
func (c *CriManager) StreamRouter() stream.Router {
	return c.StreamServer
}

// TODO: Move the underlying functions to their respective files in the future.

// Version returns the runtime name, runtime version and runtime API version.
func (c *CriManager) Version(ctx context.Context, r *runtime.VersionRequest) (*runtime.VersionResponse, error) {
	return &runtime.VersionResponse{
		Version:           kubeletRuntimeAPIVersion,
		RuntimeName:       pouchRuntimeName,
		RuntimeVersion:    version.Version,
		RuntimeApiVersion: version.APIVersion,
	}, nil
}

// RunPodSandbox creates and starts a pod-level sandbox. Runtimes should ensure
// the sandbox is in ready state.
func (c *CriManager) RunPodSandbox(ctx context.Context, r *runtime.RunPodSandboxRequest) (_ *runtime.RunPodSandboxResponse, retErr error) {
	label := util_metrics.ActionRunLabel
	defer func(start time.Time) {
		metrics.PodActionsCounter.WithLabelValues(label).Inc()
		metrics.PodActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	config := r.GetConfig()

	// Step 1: Prepare image for the sandbox.
	image := c.SandboxImage

	// Make sure the sandbox image exists.
	err := c.ensureSandboxImageExists(ctx, image)
	if err != nil {
		return nil, err
	}

	// Step 2: Create the sandbox container.

	// prepare the sandboxID and store it.
	id, err := c.generateSandboxID(ctx)
	if err != nil {
		return nil, err
	}
	sandboxMeta := &SandboxMeta{
		ID: id,
	}
	if err := c.SandboxStore.Put(sandboxMeta); err != nil {
		return nil, err
	}

	// If running sandbox failed, clean up the sandboxMeta from sandboxStore.
	// We should clean it until the container has been removed successfully by Pouchd.
	removeContainerErr := false
	defer func() {
		if retErr != nil && !removeContainerErr {
			if err := c.SandboxStore.Remove(id); err != nil {
				logrus.Errorf("failed to remove the metadata of container %q from sandboxStore: %v", id, err)
			}
		}
	}()

	// applies the runtime of container specified by the caller.
	if err := c.applySandboxRuntimeHandler(sandboxMeta, r.GetRuntimeHandler(), config.Annotations); err != nil {
		return nil, err
	}

	// applies the annotations extended.
	if err := c.applySandboxAnnotations(sandboxMeta, config.Annotations); err != nil {
		return nil, err
	}

	createConfig, err := makeSandboxPouchConfig(config, sandboxMeta.Runtime, image)

	if err != nil {
		return nil, fmt.Errorf("failed to make sandbox pouch config for pod %q: %v", config.Metadata.Name, err)
	}
	createConfig.SpecificID = id

	sandboxName := makeSandboxName(config)

	_, err = c.ContainerMgr.Create(ctx, sandboxName, createConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create a sandbox for pod %q: %v", config.Metadata.Name, err)
	}

	sandboxMeta.Config = config
	if err := c.SandboxStore.Put(sandboxMeta); err != nil {
		return nil, err
	}

	// If running sandbox failed, clean up the container.
	defer func() {
		if retErr != nil {
			if err := c.ContainerMgr.Remove(ctx, id, &apitypes.ContainerRemoveOptions{Volumes: true, Force: true}); err != nil {
				removeContainerErr = true
				logrus.Errorf("failed to remove container when running sandbox failed %q: %v", id, err)
			}
		}
	}()

	// Step 3: Start the sandbox container.
	err = c.ContainerMgr.Start(ctx, id, &apitypes.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start sandbox container for pod %q: %v", config.Metadata.Name, err)
	}

	sandboxRootDir := path.Join(c.SandboxBaseDir, id)
	err = os.MkdirAll(sandboxRootDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox root directory: %v", err)
	}
	defer func() {
		// If running sandbox failed, clean up the sandbox directory.
		if retErr != nil {
			if err := os.RemoveAll(sandboxRootDir); err != nil {
				logrus.Errorf("failed to clean up the directory of sandbox %q: %v", id, err)
			}
		}
	}()

	// Setup sandbox file /etc/resolv.conf.
	err = setupSandboxFiles(sandboxRootDir, config)
	if err != nil {
		return nil, fmt.Errorf("failed to setup sandbox files: %v", err)
	}

	// Step 4: Setup networking for the sandbox.
	networkNamespaceMode := config.GetLinux().GetSecurityContext().GetNamespaceOptions().GetNetwork()
	// If it is in host network, no need to configure the network of sandbox.
	if networkNamespaceMode != runtime.NamespaceMode_NODE {
		err = c.setupPodNetwork(ctx, id, config)
		if err != nil {
			return nil, err
		}
	}

	metrics.PodSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.RunPodSandboxResponse{PodSandboxId: id}, nil
}

// StartPodSandbox restart a sandbox pod which was stopped by accident
// and we should reconfigure it with network plugin which will make sure it reacquire its original network configuration,
// like IP address.
func (c *CriManager) StartPodSandbox(ctx context.Context, r *runtime.StartPodSandboxRequest) (*runtime.StartPodSandboxResponse, error) {
	label := util_metrics.ActionStartLabel
	defer func(start time.Time) {
		metrics.PodActionsCounter.WithLabelValues(label).Inc()
		metrics.PodActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	podSandboxID := r.GetPodSandboxId()

	// start PodSandbox.
	startErr := c.ContainerMgr.Start(ctx, podSandboxID, &apitypes.ContainerStartOptions{})
	if startErr != nil {
		return nil, fmt.Errorf("failed to start podSandbox %q: %v", podSandboxID, startErr)
	}

	var err error
	defer func() {
		if err != nil {
			stopErr := c.ContainerMgr.Stop(ctx, podSandboxID, defaultStopTimeout)
			if stopErr != nil {
				logrus.Errorf("failed to stop sandbox %q: %v", podSandboxID, stopErr)
			}
		}
	}()

	// get the sandbox's meta data.
	res, err := c.SandboxStore.Get(podSandboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata of %q from SandboxStore: %v", podSandboxID, err)
	}
	sandboxMeta := res.(*SandboxMeta)

	// setup networking for the sandbox.
	networkNamespaceMode := sandboxMeta.Config.GetLinux().GetSecurityContext().GetNamespaceOptions().GetNetwork()
	// If it is in host network, no need to configure the network of sandbox.
	if networkNamespaceMode != runtime.NamespaceMode_NODE {
		err = c.setupPodNetwork(ctx, podSandboxID, sandboxMeta.Config)
		if err != nil {
			return nil, err
		}
	}

	metrics.PodSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.StartPodSandboxResponse{}, nil
}

// StopPodSandbox stops the sandbox. If there are any running containers in the
// sandbox, they should be forcibly terminated.
func (c *CriManager) StopPodSandbox(ctx context.Context, r *runtime.StopPodSandboxRequest) (*runtime.StopPodSandboxResponse, error) {
	label := util_metrics.ActionStopLabel
	defer func(start time.Time) {
		metrics.PodActionsCounter.WithLabelValues(label).Inc()
		metrics.PodActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	podSandboxID := r.GetPodSandboxId()
	res, err := c.SandboxStore.Get(podSandboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata of %q from SandboxStore: %v", podSandboxID, err)
	}
	sandboxMeta := res.(*SandboxMeta)

	opts := &mgr.ContainerListOption{All: true}
	filter := func(c *mgr.Container) bool {
		return c.Config.Labels[sandboxIDLabelKey] == podSandboxID
	}
	opts.FilterFunc = filter

	containers, err := c.ContainerMgr.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to stop sandbox %q: %v", podSandboxID, err)
	}

	// Stop all containers in the sandbox.
	for _, container := range containers {
		err = c.ContainerMgr.Stop(ctx, container.ID, defaultStopTimeout)
		if err != nil {
			return nil, fmt.Errorf("failed to stop container %q of sandbox %q: %v", container.ID, podSandboxID, err)
		}
		logrus.Infof("success to stop container %q of sandbox %q", container.ID, podSandboxID)
	}

	container, err := c.ContainerMgr.Get(ctx, podSandboxID)
	if err != nil {
		return nil, err
	}
	metadata, err := parseSandboxName(container.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata of sandbox %q from container name: %v", podSandboxID, err)
	}

	securityContext := sandboxMeta.Config.GetLinux().GetSecurityContext()
	hostNet := securityContext.GetNamespaceOptions().GetNetwork() == runtime.NamespaceMode_NODE

	// Teardown network of the pod, if it is not in host network mode.
	if !hostNet {
		sandbox, err := c.ContainerMgr.Get(ctx, podSandboxID)
		if err != nil {
			return nil, fmt.Errorf("failed to get sandbox %q: %v", podSandboxID, err)
		}

		netNSPath := containerNetns(sandbox)
		err = c.CniMgr.TearDownPodNetwork(&ocicni.PodNetwork{
			Name:         metadata.GetName(),
			Namespace:    metadata.GetNamespace(),
			ID:           podSandboxID,
			NetNS:        netNSPath,
			PortMappings: toCNIPortMappings(sandboxMeta.Config.GetPortMappings()),
		})
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			logrus.Warnf("failed to find network namespace file %s of sandbox %s which may have been already stopped", netNSPath, podSandboxID)
		}
	}

	// Stop the sandbox container.
	err = c.ContainerMgr.Stop(ctx, podSandboxID, defaultStopTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to stop sandbox %q: %v", podSandboxID, err)
	}

	metrics.PodSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.StopPodSandboxResponse{}, nil
}

// RemovePodSandbox removes the sandbox. If there are running containers in the
// sandbox, they should be forcibly removed.
func (c *CriManager) RemovePodSandbox(ctx context.Context, r *runtime.RemovePodSandboxRequest) (*runtime.RemovePodSandboxResponse, error) {
	label := util_metrics.ActionRemoveLabel
	defer func(start time.Time) {
		metrics.PodActionsCounter.WithLabelValues(label).Inc()
		metrics.PodActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	podSandboxID := r.GetPodSandboxId()

	opts := &mgr.ContainerListOption{All: true}
	filter := func(c *mgr.Container) bool {
		return c.Config.Labels[sandboxIDLabelKey] == podSandboxID
	}
	opts.FilterFunc = filter

	containers, err := c.ContainerMgr.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to remove sandbox %q: %v", podSandboxID, err)
	}

	// Remove all containers in the sandbox.
	for _, container := range containers {
		if err := c.ContainerMgr.Remove(ctx, container.ID, &apitypes.ContainerRemoveOptions{Volumes: true, Force: true}); err != nil {
			return nil, fmt.Errorf("failed to remove container %q of sandbox %q: %v", container.ID, podSandboxID, err)
		}

		logrus.Infof("success to remove container %q of sandbox %q", container.ID, podSandboxID)
	}

	// Remove the sandbox container.
	if err := c.ContainerMgr.Remove(ctx, podSandboxID, &apitypes.ContainerRemoveOptions{Volumes: true, Force: true}); err != nil {
		return nil, fmt.Errorf("failed to remove sandbox %q: %v", podSandboxID, err)
	}

	// Cleanup the sandbox root directory.
	sandboxRootDir := path.Join(c.SandboxBaseDir, podSandboxID)

	if err := os.RemoveAll(sandboxRootDir); err != nil {
		return nil, fmt.Errorf("failed to remove root directory %q: %v", sandboxRootDir, err)
	}

	if err := c.SandboxStore.Remove(podSandboxID); err != nil {
		return nil, fmt.Errorf("failed to remove meta %q: %v", sandboxRootDir, err)
	}

	metrics.PodSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.RemovePodSandboxResponse{}, nil
}

// PodSandboxStatus returns the status of the PodSandbox.
func (c *CriManager) PodSandboxStatus(ctx context.Context, r *runtime.PodSandboxStatusRequest) (*runtime.PodSandboxStatusResponse, error) {
	label := util_metrics.ActionStatusLabel
	defer func(start time.Time) {
		metrics.PodActionsCounter.WithLabelValues(label).Inc()
		metrics.PodActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	podSandboxID := r.GetPodSandboxId()

	res, err := c.SandboxStore.Get(podSandboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata of %q from SandboxStore: %v", podSandboxID, err)
	}
	sandboxMeta := res.(*SandboxMeta)

	sandbox, err := c.ContainerMgr.Get(ctx, podSandboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status of sandbox %q: %v", podSandboxID, err)
	}

	// Parse the timestamps.
	createdAt, err := toCriTimestamp(sandbox.Created)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp for sandbox %q: %v", podSandboxID, err)
	}

	// Translate container to sandbox state.
	state := runtime.PodSandboxState_SANDBOX_NOTREADY
	if sandbox.State.Status == apitypes.StatusRunning {
		state = runtime.PodSandboxState_SANDBOX_READY
	}

	metadata, err := parseSandboxName(sandbox.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get status of sandbox %q: %v", podSandboxID, err)
	}
	labels, annotations := extractLabels(sandbox.Config.Labels)

	nsOpts := sandboxMeta.Config.GetLinux().GetSecurityContext().GetNamespaceOptions()
	hostNet := nsOpts.GetNetwork() == runtime.NamespaceMode_NODE

	var ip string
	// No need to get ip for host network mode.
	if !hostNet {
		ip, err = c.CniMgr.GetPodNetworkStatus(containerNetns(sandbox))
		if err != nil {
			// Maybe the pod has been stopped.
			logrus.Warnf("failed to get ip of sandbox %q: %v", podSandboxID, err)
		}
	}

	if v, exist := annotations[passthruKey]; exist && v == "true" {
		ip = annotations[passthruIP]
	}

	status := &runtime.PodSandboxStatus{
		Id:          sandbox.ID,
		State:       state,
		CreatedAt:   createdAt,
		Metadata:    metadata,
		Labels:      labels,
		Annotations: annotations,
		Network:     &runtime.PodSandboxNetworkStatus{Ip: ip},
		Linux: &runtime.LinuxPodSandboxStatus{
			Namespaces: &runtime.Namespace{
				Options: &runtime.NamespaceOption{
					Network: nsOpts.GetNetwork(),
					Pid:     nsOpts.GetPid(),
					Ipc:     nsOpts.GetIpc(),
				},
			},
		},
	}

	metrics.PodSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.PodSandboxStatusResponse{Status: status}, nil
}

// ListPodSandbox returns a list of Sandbox.
func (c *CriManager) ListPodSandbox(ctx context.Context, r *runtime.ListPodSandboxRequest) (*runtime.ListPodSandboxResponse, error) {
	label := util_metrics.ActionListLabel
	defer func(start time.Time) {
		metrics.PodActionsCounter.WithLabelValues(label).Inc()
		metrics.PodActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	opts := &mgr.ContainerListOption{All: true}
	filter := func(c *mgr.Container) bool {
		return c.Config.Labels[containerTypeLabelKey] == containerTypeLabelSandbox
	}
	opts.FilterFunc = filter

	// Filter *only* (sandbox) containers.
	sandboxList, err := c.ContainerMgr.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list sandbox: %v", err)
	}

	sandboxList, err = c.filterInvalidSandboxes(ctx, sandboxList)
	if err != nil {
		return nil, fmt.Errorf("failed to filter invalid sandboxes: %v", err)
	}

	sandboxes := make([]*runtime.PodSandbox, 0, len(sandboxList))
	for _, s := range sandboxList {
		sandbox, err := toCriSandbox(s)
		if err != nil {
			// TODO: log an error message?
			continue
		}
		sandboxes = append(sandboxes, sandbox)
	}

	result := filterCRISandboxes(sandboxes, r.GetFilter())

	metrics.PodSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.ListPodSandboxResponse{Items: result}, nil
}

// CreateContainer creates a new container in the given PodSandbox.
func (c *CriManager) CreateContainer(ctx context.Context, r *runtime.CreateContainerRequest) (*runtime.CreateContainerResponse, error) {
	label := util_metrics.ActionCreateLabel
	defer func(start time.Time) {
		metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	config := r.GetConfig()
	sandboxConfig := r.GetSandboxConfig()
	podSandboxID := r.GetPodSandboxId()

	// get sandbox
	sandbox, err := c.ContainerMgr.Get(ctx, podSandboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sandbox %q: %v", podSandboxID, err)
	}

	res, err := c.SandboxStore.Get(podSandboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata of %q from SandboxStore: %v", podSandboxID, err)
	}
	sandboxMeta := res.(*SandboxMeta)
	sandboxMeta.NetNS = containerNetns(sandbox)

	labels := makeLabels(config.GetLabels(), config.GetAnnotations())
	// Apply the container type lable.
	labels[containerTypeLabelKey] = containerTypeLabelContainer
	// Write the sandbox ID in the labels.
	labels[sandboxIDLabelKey] = podSandboxID
	// Get container log.
	var logPath string
	if config.GetLogPath() != "" {
		logPath = filepath.Join(sandboxConfig.GetLogDirectory(), config.GetLogPath())
		labels[containerLogPathLabelKey] = logPath
	}

	image := ""
	if iSpec := config.GetImage(); iSpec != nil {
		image = iSpec.Image
	}

	specAnnotation := make(map[string]string)
	specAnnotation[anno.ContainerType] = anno.ContainerTypeContainer
	specAnnotation[anno.SandboxName] = podSandboxID
	specAnnotation[anno.SandboxID] = podSandboxID

	resources := r.GetConfig().GetLinux().GetResources()
	createConfig := &apitypes.ContainerCreateConfig{
		ContainerConfig: apitypes.ContainerConfig{
			Entrypoint: config.Command,
			Cmd:        config.Args,
			Env:        generateEnvList(config.GetEnvs()),
			Image:      image,
			WorkingDir: config.WorkingDir,
			Labels:     labels,
			// Interactive containers:
			OpenStdin:      config.Stdin,
			StdinOnce:      config.StdinOnce,
			Tty:            config.Tty,
			SpecAnnotation: specAnnotation,
			NetPriority:    config.NetPriority,
			DiskQuota:      resources.GetDiskQuota(),
			QuotaID:        config.GetQuotaId(),
			MaskedPaths:    config.GetLinux().GetSecurityContext().GetMaskedPaths(),
			ReadonlyPaths:  config.GetLinux().GetSecurityContext().GetReadonlyPaths(),
		},
		HostConfig: &apitypes.HostConfig{
			Binds:     generateMountBindings(config.GetMounts()),
			Resources: parseResourcesFromCRI(resources),
		},
		NetworkingConfig: &apitypes.NetworkingConfig{},
	}

	err = c.updateCreateConfig(createConfig, config, sandboxConfig, sandboxMeta)
	if err != nil {
		return nil, err
	}

	// Bindings to overwrite the container's /etc/resolv.conf, /etc/hosts etc.
	sandboxRootDir := path.Join(c.SandboxBaseDir, podSandboxID)
	createConfig.HostConfig.Binds = append(createConfig.HostConfig.Binds, generateContainerMounts(sandboxRootDir)...)

	var devices []*apitypes.DeviceMapping
	for _, device := range config.Devices {
		devices = append(devices, &apitypes.DeviceMapping{
			PathOnHost:        device.HostPath,
			PathInContainer:   device.ContainerPath,
			CgroupPermissions: device.Permissions,
		})
	}
	createConfig.HostConfig.Resources.Devices = devices

	containerName := makeContainerName(sandboxConfig, config)

	// call cri plugin to update create config
	if c.CriPlugin != nil {
		if err := c.CriPlugin.PreCreateContainer(createConfig, sandboxMeta); err != nil {
			return nil, err
		}
	}

	createResp, err := c.ContainerMgr.Create(ctx, containerName, createConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create container for sandbox %q: %v", podSandboxID, err)
	}

	containerID := createResp.ID

	defer func() {
		// If the container failed to be created, clean up the container.
		if err != nil {
			removeErr := c.ContainerMgr.Remove(ctx, containerID, &apitypes.ContainerRemoveOptions{Volumes: true, Force: true})
			if removeErr != nil {
				logrus.Errorf("failed to remove the container when creating container failed: %v", removeErr)
			}
		}
	}()

	if logPath != "" {
		if err := c.ContainerMgr.AttachCRILog(ctx, containerID, logPath); err != nil {
			return nil, err
		}
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.CreateContainerResponse{ContainerId: containerID}, nil
}

// StartContainer starts the container.
func (c *CriManager) StartContainer(ctx context.Context, r *runtime.StartContainerRequest) (*runtime.StartContainerResponse, error) {
	label := util_metrics.ActionStartLabel
	defer func(start time.Time) {
		metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	containerID := r.GetContainerId()

	err := c.ContainerMgr.Start(ctx, containerID, &apitypes.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start container %q: %v", containerID, err)
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.StartContainerResponse{}, nil
}

// StopContainer stops a running container with a grace period (i.e., timeout).
func (c *CriManager) StopContainer(ctx context.Context, r *runtime.StopContainerRequest) (*runtime.StopContainerResponse, error) {
	label := util_metrics.ActionStopLabel
	defer func(start time.Time) {
		metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	containerID := r.GetContainerId()

	err := c.ContainerMgr.Stop(ctx, containerID, r.GetTimeout())
	if err != nil {
		return nil, fmt.Errorf("failed to stop container %q: %v", containerID, err)
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.StopContainerResponse{}, nil
}

// RemoveContainer removes the container.
func (c *CriManager) RemoveContainer(ctx context.Context, r *runtime.RemoveContainerRequest) (*runtime.RemoveContainerResponse, error) {
	label := util_metrics.ActionRemoveLabel
	defer func(start time.Time) {
		metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	containerID := r.GetContainerId()

	if err := c.ContainerMgr.Remove(ctx, containerID, &apitypes.ContainerRemoveOptions{Volumes: true, Force: true}); err != nil {
		return nil, fmt.Errorf("failed to remove container %q: %v", containerID, err)
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.RemoveContainerResponse{}, nil
}

// ListContainers lists all containers matching the filter.
func (c *CriManager) ListContainers(ctx context.Context, r *runtime.ListContainersRequest) (*runtime.ListContainersResponse, error) {
	label := util_metrics.ActionListLabel
	defer func(start time.Time) {
		metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	opts := &mgr.ContainerListOption{All: true}
	filter := func(c *mgr.Container) bool {
		return c.Config.Labels[containerTypeLabelKey] == containerTypeLabelContainer
	}
	opts.FilterFunc = filter

	// Filter *only* (non-sandbox) containers.
	containerList, err := c.ContainerMgr.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list container: %v", err)
	}

	containers := make([]*runtime.Container, 0, len(containerList))
	for _, c := range containerList {
		container, err := toCriContainer(c)
		if err != nil {
			logrus.Warnf("failed to translate container %v to cri container in ListContainers: %v", c.ID, err)
			continue
		}
		containers = append(containers, container)
	}

	result := filterCRIContainers(containers, r.GetFilter())

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.ListContainersResponse{Containers: result}, nil
}

// ContainerStatus inspects the container and returns the status.
func (c *CriManager) ContainerStatus(ctx context.Context, r *runtime.ContainerStatusRequest) (*runtime.ContainerStatusResponse, error) {
	label := util_metrics.ActionStatusLabel
	defer func(start time.Time) {
		metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	id := r.GetContainerId()
	container, err := c.ContainerMgr.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get container status of %q: %v", id, err)
	}

	// Parse the timestamps.
	var createdAt, startedAt, finishedAt int64
	for _, item := range []struct {
		t *int64
		s string
	}{
		{t: &createdAt, s: container.Created},
		{t: &startedAt, s: container.State.StartedAt},
		{t: &finishedAt, s: container.State.FinishedAt},
	} {
		*item.t, err = toCriTimestamp(item.s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp for container %q: %v", id, err)
		}
	}

	// Convert the mounts.
	mounts := make([]*runtime.Mount, 0, len(container.Mounts))
	for _, m := range container.Mounts {
		mounts = append(mounts, &runtime.Mount{
			HostPath:      m.Source,
			ContainerPath: m.Destination,
			Readonly:      !m.RW,
			Name:          m.Name,
			// Note: can't set SeLinuxRelabel.
		})
	}

	// Interpret container states.
	var state runtime.ContainerState
	var reason, message string
	if container.State.Status == apitypes.StatusRunning {
		// Container is running.
		state = runtime.ContainerState_CONTAINER_RUNNING
	} else {
		// Container is *not* running. We need to get more details.
		//		* Case 1: container has run and exited with non-zero finishedAt time
		//		* Case 2: container has failed to start; it has a zero finishedAt
		//				  time, but a non-zero exit code.
		//		* Case 3: container has been created, but not started (yet).
		finishTime, err := time.Parse(utils.TimeLayout, container.State.FinishedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse container finish time %s: %v", container.State.FinishedAt, err)
		}
		if !finishTime.IsZero() {
			state = runtime.ContainerState_CONTAINER_EXITED
			switch {
			case container.State.OOMKilled:
				reason = "OOMKilled"
			case container.State.ExitCode == 0:
				reason = "Completed"
			default:
				reason = "Error"
			}
		} else if container.State.ExitCode != 0 {
			state = runtime.ContainerState_CONTAINER_EXITED
			// Adjust finishedAt and startedAt time to createdAt time to avoid confusion.
			finishedAt, startedAt = createdAt, createdAt
			reason = "ContainerCannotRun"
		} else {
			state = runtime.ContainerState_CONTAINER_CREATED
		}
		message = container.State.Error
	}

	exitCode := int32(container.State.ExitCode)

	metadata, err := parseContainerName(container.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get container status of %q: %v", id, err)
	}

	labels, annotations := extractLabels(container.Config.Labels)

	// FIXME(fuwei): if user repush image with the same reference, the image
	// ID will be changed. For now, pouch daemon will remove the old image ID
	// so that CRI fails to fetch the running container. Before upgrade
	// pouch daemon image manager, we use reference to get image instead of
	// id.
	imageInfo, err := c.ImageMgr.GetImage(ctx, container.Config.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to get image %s: %v", container.Config.Image, err)
	}
	imageRef := imageInfo.ID
	if len(imageInfo.RepoDigests) > 0 {
		imageRef = imageInfo.RepoDigests[0]
	}

	logPath := labels[containerLogPathLabelKey]

	resources := container.HostConfig.Resources
	diskQuota := container.Config.DiskQuota
	status := &runtime.ContainerStatus{
		Id:          container.ID,
		Metadata:    metadata,
		Image:       &runtime.ImageSpec{Image: container.Config.Image},
		ImageRef:    imageRef,
		Mounts:      mounts,
		ExitCode:    exitCode,
		State:       state,
		CreatedAt:   createdAt,
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
		Reason:      reason,
		Message:     message,
		Labels:      labels,
		Annotations: annotations,
		LogPath:     logPath,
		Volumes:     parseVolumesFromPouch(container.Config.Volumes),
		Resources:   parseResourcesFromPouch(resources, diskQuota),
		QuotaId:     container.Config.QuotaID,
		Envs:        parseEnvsFromPouch(container.Config.Env),
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.ContainerStatusResponse{Status: status}, nil
}

// ContainerStats returns stats of the container. If the container does not
// exist, the call returns an error.
func (c *CriManager) ContainerStats(ctx context.Context, r *runtime.ContainerStatsRequest) (*runtime.ContainerStatsResponse, error) {
	label := util_metrics.ActionStatsLabel
	defer func(start time.Time) {
		metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	containerID := r.GetContainerId()

	container, err := c.ContainerMgr.Get(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container %q with error: %v", containerID, err)
	}

	cs, err := c.getContainerMetrics(ctx, container)
	if err != nil {
		return nil, fmt.Errorf("failed to decode container metrics: %v", err)
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.ContainerStatsResponse{Stats: cs}, nil
}

// ListContainerStats returns stats of all running containers.
func (c *CriManager) ListContainerStats(ctx context.Context, r *runtime.ListContainerStatsRequest) (*runtime.ListContainerStatsResponse, error) {
	label := util_metrics.ActionStatsListLabel
	defer func(start time.Time) {
		metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	opts := &mgr.ContainerListOption{All: true}
	filter := func(c *mgr.Container) bool {
		if c.Config.Labels[containerTypeLabelKey] != containerTypeLabelContainer {
			return false
		}

		if r.GetFilter().GetId() != "" && c.ID != r.GetFilter().GetId() {
			return false
		}
		if r.GetFilter().GetPodSandboxId() != "" && c.Config.Labels[sandboxIDLabelKey] != r.GetFilter().GetPodSandboxId() {
			return false
		}
		if r.GetFilter().GetLabelSelector() != nil &&
			!criutils.MatchLabelSelector(r.GetFilter().GetLabelSelector(), c.Config.Labels) {
			return false
		}
		return true
	}
	opts.FilterFunc = filter

	containers, err := c.ContainerMgr.List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	result := &runtime.ListContainerStatsResponse{}
	for _, container := range containers {
		cs, err := c.getContainerMetrics(ctx, container)
		if err != nil {
			logrus.Warnf("failed to decode metrics of container %q: %v", container.ID, err)
			continue
		}

		result.Stats = append(result.Stats, cs)
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	return result, nil
}

// UpdateContainerResources updates ContainerConfig of the container.
func (c *CriManager) UpdateContainerResources(ctx context.Context, r *runtime.UpdateContainerResourcesRequest) (*runtime.UpdateContainerResourcesResponse, error) {
	label := util_metrics.ActionUpdateLabel
	defer func(start time.Time) {
		metrics.ContainerActionsCounter.WithLabelValues(label).Inc()
		metrics.ContainerActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	containerID := r.GetContainerId()
	container, err := c.ContainerMgr.Get(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container %q: %v", containerID, err)
	}

	// cannot update container resource when it is in removing state
	if container.IsRemoving() {
		return nil, fmt.Errorf("cannot to update resource for container %q when it is in removing state", containerID)
	}

	resources := r.GetLinux()
	updateConfig := &apitypes.UpdateConfig{
		Resources: parseResourcesFromCRI(resources),
		DiskQuota: resources.GetDiskQuota(),
	}
	err = c.ContainerMgr.Update(ctx, containerID, updateConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to update resource for container %q: %v", containerID, err)
	}

	metrics.ContainerSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.UpdateContainerResourcesResponse{}, nil
}

// ReopenContainerLog asks runtime to reopen the stdout/stderr log file
// for the container. This is often called after the log file has been
// rotated. If the container is not running, container runtime can choose
// to either create a new log file and return nil, or return an error.
// Once it returns error, new container log file MUST NOT be created.
func (c *CriManager) ReopenContainerLog(ctx context.Context, r *runtime.ReopenContainerLogRequest) (*runtime.ReopenContainerLogResponse, error) {
	containerID := r.GetContainerId()

	container, err := c.ContainerMgr.Get(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container %q with error: %v", containerID, err)
	}
	if !container.IsRunning() {
		return nil, errors.Wrap(errtypes.ErrPreCheckFailed, "container is not running")
	}

	// get logPath of container
	logPath := container.Config.Labels[containerLogPathLabelKey]
	if logPath == "" {
		logrus.Warnf("log path of container: %q is empty", containerID)
		return &runtime.ReopenContainerLogResponse{}, nil
	}

	if err := c.ContainerMgr.AttachCRILog(ctx, container.Name, logPath); err != nil {
		return nil, err
	}

	return &runtime.ReopenContainerLogResponse{}, nil
}

// ExecSync executes a command in the container, and returns the stdout output.
// If command exits with a non-zero exit code, an error is returned.
func (c *CriManager) ExecSync(ctx context.Context, r *runtime.ExecSyncRequest) (*runtime.ExecSyncResponse, error) {
	id := r.GetContainerId()

	timeout := time.Duration(r.GetTimeout()) * time.Second
	var cancel context.CancelFunc
	if timeout == 0 {
		ctx, cancel = context.WithCancel(ctx)
	} else {
		ctx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	createConfig := &apitypes.ExecCreateConfig{
		Cmd: r.GetCmd(),
	}
	execid, err := c.ContainerMgr.CreateExec(ctx, id, createConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec for container %q: %v", id, err)
	}

	stdoutBuf, stderrBuf := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	attachCfg := &pkgstreams.AttachConfig{
		UseStdout: true,
		Stdout:    stdoutBuf,
		UseStderr: true,
		Stderr:    stderrBuf,
	}
	if err := c.ContainerMgr.StartExec(ctx, execid, attachCfg); err != nil {
		return nil, fmt.Errorf("failed to start exec for container %q: %v", id, err)
	}

	execConfig, err := c.ContainerMgr.GetExecConfig(ctx, execid)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect exec for container %q: %v", id, err)
	}

	return &runtime.ExecSyncResponse{
		Stdout:   stdoutBuf.Bytes(),
		Stderr:   stderrBuf.Bytes(),
		ExitCode: int32(execConfig.ExitCode),
	}, nil
}

// Exec prepares a streaming endpoint to execute a command in the container, and returns the address.
func (c *CriManager) Exec(ctx context.Context, r *runtime.ExecRequest) (*runtime.ExecResponse, error) {
	return c.StreamServer.GetExec(r)
}

// Attach prepares a streaming endpoint to attach to a running container, and returns the address.
func (c *CriManager) Attach(ctx context.Context, r *runtime.AttachRequest) (*runtime.AttachResponse, error) {
	return c.StreamServer.GetAttach(r)
}

// PortForward prepares a streaming endpoint to forward ports from a PodSandbox, and returns the address.
func (c *CriManager) PortForward(ctx context.Context, r *runtime.PortForwardRequest) (*runtime.PortForwardResponse, error) {
	return c.StreamServer.GetPortForward(r)
}

// UpdateRuntimeConfig updates the runtime config. Currently only handles podCIDR updates.
func (c *CriManager) UpdateRuntimeConfig(ctx context.Context, r *runtime.UpdateRuntimeConfigRequest) (*runtime.UpdateRuntimeConfigResponse, error) {
	return nil, fmt.Errorf("UpdateRuntimeConfig Not Implemented Yet")
}

// Status returns the status of the runtime.
func (c *CriManager) Status(ctx context.Context, r *runtime.StatusRequest) (*runtime.StatusResponse, error) {
	label := util_metrics.ActionStatusLabel
	// record the time spent during image pull procedure.
	defer func(start time.Time) {
		metrics.RuntimeActionsCounter.WithLabelValues(label).Inc()
		metrics.RuntimeActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	runtimeCondition := &runtime.RuntimeCondition{
		Type:   runtime.RuntimeReady,
		Status: true,
	}
	networkCondition := &runtime.RuntimeCondition{
		Type:   runtime.NetworkReady,
		Status: true,
	}

	// Check the status of the cni initialization
	if err := c.CniMgr.Status(); err != nil {
		networkCondition.Status = false
		networkCondition.Reason = networkNotReadyReason
		networkCondition.Message = fmt.Sprintf("Network plugin returns error: %v", err)
	}

	resp := &runtime.StatusResponse{
		Status: &runtime.RuntimeStatus{Conditions: []*runtime.RuntimeCondition{
			runtimeCondition,
			networkCondition,
		}},
	}

	if r.Verbose {
		resp.Info = make(map[string]string)
		versionByt, err := json.Marshal(goruntime.Version())
		if err != nil {
			return nil, err
		}
		resp.Info["golang"] = string(versionByt)

		// TODO return more info
	}

	metrics.RuntimeSuccessActionsCounter.WithLabelValues(label).Inc()

	return resp, nil
}

// ListImages lists existing images.
func (c *CriManager) ListImages(ctx context.Context, r *runtime.ListImagesRequest) (*runtime.ListImagesResponse, error) {
	label := util_metrics.ActionListLabel
	// record the time spent during image pull procedure.
	defer func(start time.Time) {
		metrics.ImageActionsCounter.WithLabelValues(label).Inc()
		metrics.ImageActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	// TODO: handle image list filters.
	imageList, err := c.ImageMgr.ListImages(ctx, filters.NewArgs())
	if err != nil {
		return nil, err
	}

	// We may get images with same id and different repoTag or repoDigest,
	// so we need idExist to de-dup.
	idExist := make(map[string]bool)

	images := make([]*runtime.Image, 0, len(imageList))
	for _, i := range imageList {
		if _, ok := idExist[i.ID]; ok {
			continue
		}
		// NOTE: we should query image cache to get the correct image info.
		imageInfo, err := c.ImageMgr.GetImage(ctx, i.ID)
		if err != nil {
			continue
		}
		image, err := imageToCriImage(imageInfo)
		if err != nil {
			// TODO: log an error message?
			continue
		}
		images = append(images, image)
		idExist[i.ID] = true
	}

	metrics.ImageSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.ListImagesResponse{Images: images}, nil
}

// ImageStatus returns the status of the image. If the image is not present,
// returns a response with ImageStatusResponse.Image set to nil.
func (c *CriManager) ImageStatus(ctx context.Context, r *runtime.ImageStatusRequest) (*runtime.ImageStatusResponse, error) {
	label := util_metrics.ActionStatusLabel
	defer func(start time.Time) {
		metrics.ImageActionsCounter.WithLabelValues(label).Inc()
		metrics.ImageActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	imageRef := r.GetImage().GetImage()
	ref, err := reference.Parse(imageRef)
	if err != nil {
		return nil, err
	}

	imageInfo, err := c.ImageMgr.GetImage(ctx, ref.String())
	if err != nil {
		if errtypes.IsNotfound(err) {
			return &runtime.ImageStatusResponse{}, nil
		}
		return nil, err
	}

	image, err := imageToCriImage(imageInfo)
	if err != nil {
		return nil, err
	}

	metrics.ImageSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.ImageStatusResponse{Image: image}, nil
}

// PullImage pulls an image with authentication config.
func (c *CriManager) PullImage(ctx context.Context, r *runtime.PullImageRequest) (*runtime.PullImageResponse, error) {
	// TODO: authentication.
	imageRef := r.GetImage().GetImage()

	label := util_metrics.ActionPullLabel
	// record the time spent during image pull procedure.
	defer func(start time.Time) {
		metrics.ImageActionsCounter.WithLabelValues(label).Inc()
		metrics.ImagePullSummary.WithLabelValues(imageRef).Observe(util_metrics.SinceInMicroseconds(start))
		metrics.ImageActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	authConfig := &apitypes.AuthConfig{}
	if r.Auth != nil {
		authConfig.Auth = r.Auth.Auth
		authConfig.Username = r.Auth.Username
		authConfig.Password = r.Auth.Password
		authConfig.ServerAddress = r.Auth.ServerAddress
		authConfig.IdentityToken = r.Auth.IdentityToken
		authConfig.RegistryToken = r.Auth.RegistryToken
	}

	if err := c.ImageMgr.PullImage(ctx, imageRef, authConfig, bytes.NewBuffer([]byte{})); err != nil {
		return nil, err
	}

	imageInfo, err := c.ImageMgr.GetImage(ctx, imageRef)
	if err != nil {
		return nil, err
	}

	metrics.ImageSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.PullImageResponse{ImageRef: imageInfo.ID}, nil
}

// RemoveImage removes the image.
func (c *CriManager) RemoveImage(ctx context.Context, r *runtime.RemoveImageRequest) (*runtime.RemoveImageResponse, error) {
	label := util_metrics.ActionRemoveLabel
	defer func(start time.Time) {
		metrics.ImageActionsCounter.WithLabelValues(label).Inc()
		metrics.ImageActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	imageRef := r.GetImage().GetImage()

	if err := c.ImageMgr.RemoveImage(ctx, imageRef, false); err != nil {
		if errtypes.IsNotfound(err) {
			// Now we just return empty if the ErrorNotFound occurred.
			return &runtime.RemoveImageResponse{}, nil
		}
		return nil, err
	}

	metrics.ImageSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.RemoveImageResponse{}, nil
}

// ImageFsInfo returns information of the filesystem that is used to store images.
func (c *CriManager) ImageFsInfo(ctx context.Context, r *runtime.ImageFsInfoRequest) (*runtime.ImageFsInfoResponse, error) {
	label := util_metrics.ActionInfoLabel
	defer func(start time.Time) {
		metrics.ImageActionsCounter.WithLabelValues(label).Inc()
		metrics.ImageActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	snapshots := c.SnapshotStore.List()
	timestamp := time.Now().UnixNano()
	var usedBytes, inodesUsed uint64
	for _, sn := range snapshots {
		// Use the oldest timestamp as the timestamp of imagefs info.
		if sn.Timestamp < timestamp {
			timestamp = sn.Timestamp
		}
		usedBytes += sn.Size
		inodesUsed += sn.Inodes
	}

	metrics.ImageSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.ImageFsInfoResponse{
		ImageFilesystems: []*runtime.FilesystemUsage{
			{
				Timestamp:  timestamp,
				FsId:       &runtime.FilesystemIdentifier{Mountpoint: c.imageFSPath},
				UsedBytes:  &runtime.UInt64Value{Value: usedBytes},
				InodesUsed: &runtime.UInt64Value{Value: inodesUsed},
			},
		},
	}, nil
}

// RemoveVolume removes the volume.
func (c *CriManager) RemoveVolume(ctx context.Context, r *runtime.RemoveVolumeRequest) (*runtime.RemoveVolumeResponse, error) {
	label := util_metrics.ActionRemoveLabel
	defer func(start time.Time) {
		metrics.VolumeActionsCounter.WithLabelValues(label).Inc()
		metrics.VolumeActionsTimer.WithLabelValues(label).Observe(time.Since(start).Seconds())
	}(time.Now())

	volumeName := r.GetVolumeName()
	if err := c.VolumeMgr.Remove(ctx, volumeName); err != nil {
		return nil, err
	}

	metrics.VolumeSuccessActionsCounter.WithLabelValues(label).Inc()

	return &runtime.RemoveVolumeResponse{}, nil
}
