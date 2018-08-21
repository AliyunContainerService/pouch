package v1alpha1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	goruntime "runtime"
	"time"

	apitypes "github.com/alibaba/pouch/apis/types"
	anno "github.com/alibaba/pouch/cri/annotations"
	cni "github.com/alibaba/pouch/cri/ocicni"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/meta"
	"github.com/alibaba/pouch/pkg/reference"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/version"

	// NOTE: "golang.org/x/net/context" is compatible with standard "context" in golang1.7+.
	"github.com/cri-o/ocicni/pkg/ocicni"
	"github.com/sirupsen/logrus"
	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
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

	// sandboxContainerName is a string to include in the pouch container so
	// that users can easily identify the sandboxes.
	sandboxContainerName = "POD"

	// nameDelimiter is used to construct pouch container names.
	nameDelimiter = "_"

	// Address and port of stream server.
	// TODO: specify them in the parameters of pouchd.
	streamServerAddress = ""
	streamServerPort    = "10010"

	namespaceModeHost = "host"
	namespaceModeNone = "none"

	// resolvConfPath is the abs path of resolv.conf on host or container.
	resolvConfPath = "/etc/resolv.conf"

	// statsCollectPeriod is the time duration (in time.Second)  we sync stats from containerd.
	statsCollectPeriod = 10

	// defaultSnapshotterName is the default Snapshotter name.
	defaultSnapshotterName = "overlayfs"

	// snapshotPlugin implements a snapshotter.
	snapshotPlugin = "io.containerd.snapshotter.v1"

	// networkNotReadyReason is the reason reported when network is not ready.
	networkNotReadyReason = "NetworkPluginNotReady"
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

	// StreamServerStart starts the stream server of CRI.
	StreamServerStart() error
}

// CriManager is an implementation of interface CriMgr.
type CriManager struct {
	ContainerMgr mgr.ContainerMgr
	ImageMgr     mgr.ImageMgr
	CniMgr       cni.CniMgr

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

	// ImageFSUUID is the device uuid of image filesystem.
	ImageFSUUID string
}

// NewCriManager creates a brand new cri manager.
func NewCriManager(config *config.Config, ctrMgr mgr.ContainerMgr, imgMgr mgr.ImageMgr) (CriMgr, error) {
	streamServer, err := newStreamServer(ctrMgr, streamServerAddress, streamServerPort)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream server for cri manager: %v", err)
	}

	c := &CriManager{
		ContainerMgr:   ctrMgr,
		ImageMgr:       imgMgr,
		CniMgr:         cni.NewCniManager(&config.CriConfig),
		StreamServer:   streamServer,
		SandboxBaseDir: path.Join(config.HomeDir, "sandboxes"),
		SandboxImage:   config.CriConfig.SandboxImage,
		SnapshotStore:  mgr.NewSnapshotStore(),
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

	imageFSPath := imageFSPath(path.Join(config.HomeDir, "containerd/root"), defaultSnapshotterName)
	c.ImageFSUUID, err = getDeviceUUID(imageFSPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get imagefs uuid of %q: %v", imageFSPath, err)
	}

	snapshotsSyncer := ctrMgr.NewSnapshotsSyncer(
		c.SnapshotStore,
		time.Duration(statsCollectPeriod)*time.Second,
	)
	snapshotsSyncer.Start()

	return NewCriWrapper(c), nil
}

// StreamServerStart starts the stream server of CRI.
func (c *CriManager) StreamServerStart() error {
	return c.StreamServer.Start()
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
	config := r.GetConfig()

	// Step 1: Prepare image for the sandbox.
	image := c.SandboxImage

	// Make sure the sandbox image exists.
	err := c.ensureSandboxImageExists(ctx, image)
	if err != nil {
		return nil, err
	}

	// Step 2: Create the sandbox container.
	createConfig, err := makeSandboxPouchConfig(config, image)
	if err != nil {
		return nil, fmt.Errorf("failed to make sandbox pouch config for pod %q: %v", config.Metadata.Name, err)
	}

	sandboxName := makeSandboxName(config)

	createResp, err := c.ContainerMgr.Create(ctx, sandboxName, createConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create a sandbox for pod %q: %v", config.Metadata.Name, err)
	}
	id := createResp.ID
	defer func() {
		// If running sandbox failed, clean up the container.
		if retErr != nil {
			err := c.ContainerMgr.Remove(ctx, id, &apitypes.ContainerRemoveOptions{Volumes: true, Force: true})
			if err != nil {
				logrus.Errorf("failed to remove the container when running sandbox failed: %v", err)
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
			os.RemoveAll(sandboxRootDir)
		}
	}()

	// Setup sandbox file /etc/resolv.conf.
	err = setupSandboxFiles(sandboxRootDir, config)
	if err != nil {
		return nil, fmt.Errorf("failed to setup sandbox files: %v", err)
	}

	// Step 4: Setup networking for the sandbox.
	var netnsPath string
	securityContext := config.GetLinux().GetSecurityContext()
	hostNet := securityContext.GetNamespaceOptions().GetHostNetwork()
	// If it is in host network, no need to configure the network of sandbox.
	if !hostNet {
		container, err := c.ContainerMgr.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		netnsPath = containerNetns(container)
		if netnsPath == "" {
			return nil, fmt.Errorf("failed to find network namespace path for sandbox %q", id)
		}

		err = c.CniMgr.SetUpPodNetwork(&ocicni.PodNetwork{
			Name:         config.GetMetadata().GetName(),
			Namespace:    config.GetMetadata().GetNamespace(),
			ID:           id,
			NetNS:        netnsPath,
			PortMappings: toCNIPortMappings(config.GetPortMappings()),
		})
		if err != nil {
			return nil, err
		}
	}

	sandboxMeta := &SandboxMeta{
		ID:        id,
		Config:    config,
		NetNSPath: netnsPath,
		Runtime:   config.Annotations[anno.KubernetesRuntime],
	}
	c.SandboxStore.Put(sandboxMeta)

	return &runtime.RunPodSandboxResponse{PodSandboxId: id}, nil
}

// StopPodSandbox stops the sandbox. If there are any running containers in the
// sandbox, they should be forcibly terminated.
func (c *CriManager) StopPodSandbox(ctx context.Context, r *runtime.StopPodSandboxRequest) (*runtime.StopPodSandboxResponse, error) {
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
	hostNet := securityContext.GetNamespaceOptions().GetHostNetwork()

	// Teardown network of the pod, if it is not in host network mode.
	if !hostNet {
		_, err = os.Stat(sandboxMeta.NetNSPath)
		// If the sandbox has been stopped, the corresponding network namespace will not exist.
		if err == nil {
			err = c.CniMgr.TearDownPodNetwork(&ocicni.PodNetwork{
				Name:         metadata.GetName(),
				Namespace:    metadata.GetNamespace(),
				ID:           podSandboxID,
				NetNS:        sandboxMeta.NetNSPath,
				PortMappings: toCNIPortMappings(sandboxMeta.Config.GetPortMappings()),
			})
			if err != nil {
				return nil, err
			}
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to stat network namespace file %s of sandbox %s: %v", sandboxMeta.NetNSPath, podSandboxID, err)
		}
	}

	// Stop the sandbox container.
	err = c.ContainerMgr.Stop(ctx, podSandboxID, defaultStopTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to stop sandbox %q: %v", podSandboxID, err)
	}

	return &runtime.StopPodSandboxResponse{}, nil
}

// RemovePodSandbox removes the sandbox. If there are running containers in the
// sandbox, they should be forcibly removed.
func (c *CriManager) RemovePodSandbox(ctx context.Context, r *runtime.RemovePodSandboxRequest) (*runtime.RemovePodSandboxResponse, error) {
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
		err = c.ContainerMgr.Remove(ctx, container.ID, &apitypes.ContainerRemoveOptions{Volumes: true, Force: true})
		if err != nil {
			return nil, fmt.Errorf("failed to remove container %q of sandbox %q: %v", container.ID, podSandboxID, err)
		}
	}

	// Remove the sandbox container.
	err = c.ContainerMgr.Remove(ctx, podSandboxID, &apitypes.ContainerRemoveOptions{Volumes: true, Force: true})
	if err != nil {
		return nil, fmt.Errorf("failed to remove sandbox %q: %v", podSandboxID, err)
	}

	// Cleanup the sandbox root directory.
	sandboxRootDir := path.Join(c.SandboxBaseDir, podSandboxID)
	err = os.RemoveAll(sandboxRootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to remove root directory %q: %v", sandboxRootDir, err)
	}

	err = c.SandboxStore.Remove(podSandboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to remove meta %q: %v", sandboxRootDir, err)
	}

	return &runtime.RemovePodSandboxResponse{}, nil
}

// PodSandboxStatus returns the status of the PodSandbox.
func (c *CriManager) PodSandboxStatus(ctx context.Context, r *runtime.PodSandboxStatusRequest) (*runtime.PodSandboxStatusResponse, error) {
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

	securityContext := sandboxMeta.Config.GetLinux().GetSecurityContext()
	hostNet := securityContext.GetNamespaceOptions().GetHostNetwork()

	var ip string
	// No need to get ip for host network mode.
	if !hostNet {
		ip, err = c.CniMgr.GetPodNetworkStatus(sandboxMeta.NetNSPath)
		if err != nil {
			// Maybe the pod has been stopped.
			logrus.Warnf("failed to get ip of sandbox %q: %v", podSandboxID, err)
		}
	}

	status := &runtime.PodSandboxStatus{
		Id:          sandbox.ID,
		State:       state,
		CreatedAt:   createdAt,
		Metadata:    metadata,
		Labels:      labels,
		Annotations: annotations,
		Network:     &runtime.PodSandboxNetworkStatus{Ip: ip},
		// TODO: linux specific pod status.
	}

	return &runtime.PodSandboxStatusResponse{Status: status}, nil
}

// ListPodSandbox returns a list of Sandbox.
func (c *CriManager) ListPodSandbox(ctx context.Context, r *runtime.ListPodSandboxRequest) (*runtime.ListPodSandboxResponse, error) {
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

	return &runtime.ListPodSandboxResponse{Items: result}, nil
}

// CreateContainer creates a new container in the given PodSandbox.
func (c *CriManager) CreateContainer(ctx context.Context, r *runtime.CreateContainerRequest) (*runtime.CreateContainerResponse, error) {
	config := r.GetConfig()
	sandboxConfig := r.GetSandboxConfig()
	podSandboxID := r.GetPodSandboxId()

	labels := makeLabels(config.GetLabels(), config.GetAnnotations())
	// Apply the container type lable.
	labels[containerTypeLabelKey] = containerTypeLabelContainer
	// Write the sandbox ID in the labels.
	labels[sandboxIDLabelKey] = podSandboxID

	image := ""
	if iSpec := config.GetImage(); iSpec != nil {
		image = iSpec.Image
	}

	specAnnotation := make(map[string]string)
	specAnnotation[anno.ContainerType] = anno.ContainerTypeContainer
	specAnnotation[anno.SandboxName] = podSandboxID

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
		},
		HostConfig: &apitypes.HostConfig{
			Binds: generateMountBindings(config.GetMounts()),
		},
		NetworkingConfig: &apitypes.NetworkingConfig{},
	}
	err := c.updateCreateConfig(createConfig, config, sandboxConfig, podSandboxID)
	if err != nil {
		return nil, err
	}

	// Bindings to overwrite the container's /etc/resolv.conf, /etc/hosts etc.
	sandboxRootDir := path.Join(c.SandboxBaseDir, podSandboxID)
	createConfig.HostConfig.Binds = append(createConfig.HostConfig.Binds, generateContainerMounts(sandboxRootDir)...)

	// TODO: devices and security option configurations.

	containerName := makeContainerName(sandboxConfig, config)

	createResp, err := c.ContainerMgr.Create(ctx, containerName, createConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create container for sandbox %q: %v", podSandboxID, err)
	}

	containerID := createResp.ID

	// Get container log.
	if config.GetLogPath() != "" {
		logPath := filepath.Join(sandboxConfig.GetLogDirectory(), config.GetLogPath())
		// NOTE: If we attach log here, the IO of container will be created
		// by this function first, so we should decide whether open the stdin
		// here. It's weird actually, make it more elegant in the future.
		err := c.attachLog(logPath, containerID, config.Stdin)
		if err != nil {
			return nil, err
		}
	}

	return &runtime.CreateContainerResponse{ContainerId: containerID}, nil
}

// StartContainer starts the container.
func (c *CriManager) StartContainer(ctx context.Context, r *runtime.StartContainerRequest) (*runtime.StartContainerResponse, error) {
	containerID := r.GetContainerId()

	err := c.ContainerMgr.Start(ctx, containerID, &apitypes.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start container %q: %v", containerID, err)
	}

	return &runtime.StartContainerResponse{}, nil
}

// StopContainer stops a running container with a grace period (i.e., timeout).
func (c *CriManager) StopContainer(ctx context.Context, r *runtime.StopContainerRequest) (*runtime.StopContainerResponse, error) {
	containerID := r.GetContainerId()

	err := c.ContainerMgr.Stop(ctx, containerID, r.GetTimeout())
	if err != nil {
		return nil, fmt.Errorf("failed to stop container %q: %v", containerID, err)
	}

	return &runtime.StopContainerResponse{}, nil
}

// RemoveContainer removes the container.
func (c *CriManager) RemoveContainer(ctx context.Context, r *runtime.RemoveContainerRequest) (*runtime.RemoveContainerResponse, error) {
	containerID := r.GetContainerId()

	err := c.ContainerMgr.Remove(ctx, containerID, &apitypes.ContainerRemoveOptions{Volumes: true, Force: true})
	if err != nil {
		return nil, fmt.Errorf("failed to remove container %q: %v", containerID, err)
	}

	return &runtime.RemoveContainerResponse{}, nil
}

// ListContainers lists all containers matching the filter.
func (c *CriManager) ListContainers(ctx context.Context, r *runtime.ListContainersRequest) (*runtime.ListContainersResponse, error) {
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
			// TODO: log an error message?
			continue
		}
		containers = append(containers, container)
	}

	result := filterCRIContainers(containers, r.GetFilter())

	return &runtime.ListContainersResponse{Containers: result}, nil
}

// ContainerStatus inspects the container and returns the status.
func (c *CriManager) ContainerStatus(ctx context.Context, r *runtime.ContainerStatusRequest) (*runtime.ContainerStatusResponse, error) {
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

	podSandboxID := container.Config.Labels[sandboxIDLabelKey]
	res, err := c.SandboxStore.Get(podSandboxID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata of %q from SandboxStore: %v", podSandboxID, err)
	}
	sandboxMeta := res.(*SandboxMeta)
	logDirectory := sandboxMeta.Config.GetLogDirectory()
	// TODO: let the container manager handle the log stuff for CRI.
	logPath := makeupLogPath(logDirectory, metadata)

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
	}

	return &runtime.ContainerStatusResponse{Status: status}, nil
}

// ContainerStats returns stats of the container. If the container does not
// exist, the call returns an error.
func (c *CriManager) ContainerStats(ctx context.Context, r *runtime.ContainerStatsRequest) (*runtime.ContainerStatsResponse, error) {
	containerID := r.GetContainerId()

	container, err := c.ContainerMgr.Get(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container %q with error: %v", containerID, err)
	}

	cs, err := c.getContainerMetrics(ctx, container)
	if err != nil {
		return nil, fmt.Errorf("failed to decode container metrics: %v", err)
	}

	return &runtime.ContainerStatsResponse{Stats: cs}, nil
}

// ListContainerStats returns stats of all running containers.
func (c *CriManager) ListContainerStats(ctx context.Context, r *runtime.ListContainerStatsRequest) (*runtime.ListContainerStatsResponse, error) {
	opts := &mgr.ContainerListOption{All: true}
	filter := func(c *mgr.Container) bool {
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
			logrus.Errorf("failed to decode metrics of container %q: %v", container.ID, err)
			continue
		}

		result.Stats = append(result.Stats, cs)
	}

	return result, nil
}

// UpdateContainerResources updates ContainerConfig of the container.
func (c *CriManager) UpdateContainerResources(ctx context.Context, r *runtime.UpdateContainerResourcesRequest) (*runtime.UpdateContainerResourcesResponse, error) {
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
	}
	err = c.ContainerMgr.Update(ctx, containerID, updateConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to update resource for container %q: %v", containerID, err)
	}

	return &runtime.UpdateContainerResourcesResponse{}, nil
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

	reader, writer := io.Pipe()
	defer writer.Close()

	attachConfig := &mgr.AttachConfig{
		Stdout:      true,
		Stderr:      true,
		Pipe:        writer,
		MuxDisabled: true,
	}

	err = c.ContainerMgr.StartExec(ctx, execid, attachConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to start exec for container %q: %v", id, err)
	}

	readWaitCh := make(chan error, 1)
	var recv bytes.Buffer
	go func() {
		defer reader.Close()
		_, err = io.Copy(&recv, reader)
		readWaitCh <- err
	}()

	select {
	case <-ctx.Done():
		//TODO maybe stop the execution?
		return nil, fmt.Errorf("timeout %v exceeded", timeout)
	case readWaitErr := <-readWaitCh:
		if readWaitErr != nil {
			return nil, fmt.Errorf("failed to read data from the pipe: %v", err)
		}
		execConfig, err := c.ContainerMgr.GetExecConfig(ctx, execid)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect exec for container %q: %v", id, err)
		}

		var stderr []byte
		if execConfig.Error != nil {
			stderr = []byte(execConfig.Error.Error())
		}

		return &runtime.ExecSyncResponse{
			Stdout:   recv.Bytes(),
			Stderr:   stderr,
			ExitCode: int32(execConfig.ExitCode),
		}, nil
	}
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

	return resp, nil
}

// ListImages lists existing images.
func (c *CriManager) ListImages(ctx context.Context, r *runtime.ListImagesRequest) (*runtime.ListImagesResponse, error) {
	// TODO: handle image list filters.
	imageList, err := c.ImageMgr.ListImages(ctx, "")
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

	return &runtime.ListImagesResponse{Images: images}, nil
}

// ImageStatus returns the status of the image, returns nil if the image isn't present.
func (c *CriManager) ImageStatus(ctx context.Context, r *runtime.ImageStatusRequest) (*runtime.ImageStatusResponse, error) {
	imageRef := r.GetImage().GetImage()
	ref, err := reference.Parse(imageRef)
	if err != nil {
		return nil, err
	}

	imageInfo, err := c.ImageMgr.GetImage(ctx, ref.String())
	if err != nil {
		// TODO: separate ErrImageNotFound with others.
		// Now we just return empty if the error occurred.
		return &runtime.ImageStatusResponse{}, nil
	}

	image, err := imageToCriImage(imageInfo)
	if err != nil {
		return nil, err
	}

	return &runtime.ImageStatusResponse{Image: image}, nil
}

// PullImage pulls an image with authentication config.
func (c *CriManager) PullImage(ctx context.Context, r *runtime.PullImageRequest) (*runtime.PullImageResponse, error) {
	// TODO: authentication.
	imageRef := r.GetImage().GetImage()

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

	return &runtime.PullImageResponse{ImageRef: imageInfo.ID}, nil
}

// RemoveImage removes the image.
func (c *CriManager) RemoveImage(ctx context.Context, r *runtime.RemoveImageRequest) (*runtime.RemoveImageResponse, error) {
	imageRef := r.GetImage().GetImage()

	if err := c.ImageMgr.RemoveImage(ctx, imageRef, false); err != nil {
		if errtypes.IsNotfound(err) {
			// TODO: separate ErrImageNotFound with others.
			// Now we just return empty if the error occurred.
			return &runtime.RemoveImageResponse{}, nil
		}
		return nil, err
	}
	return &runtime.RemoveImageResponse{}, nil
}

// ImageFsInfo returns information of the filesystem that is used to store images.
func (c *CriManager) ImageFsInfo(ctx context.Context, r *runtime.ImageFsInfoRequest) (*runtime.ImageFsInfoResponse, error) {
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
	return &runtime.ImageFsInfoResponse{
		ImageFilesystems: []*runtime.FilesystemUsage{
			{
				Timestamp:  timestamp,
				StorageId:  &runtime.StorageIdentifier{Uuid: c.ImageFSUUID},
				UsedBytes:  &runtime.UInt64Value{Value: usedBytes},
				InodesUsed: &runtime.UInt64Value{Value: inodesUsed},
			},
		},
	}, nil
}
