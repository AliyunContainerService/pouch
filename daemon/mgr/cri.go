package mgr

import (
	"bytes"
	"fmt"

	apitypes "github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/reference"
	"github.com/alibaba/pouch/version"

	// NOTE: "golang.org/x/net/context" is compatible with standard "context" in golang1.7+.
	"golang.org/x/net/context"
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

	defaultSandboxImage = "k8s.gcr.io/pause-amd64:3.0"
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
}

// CriManager is an implementation of interface CriMgr.
type CriManager struct {
	ContainerMgr ContainerMgr
	ImageMgr     ImageMgr
}

// NewCriManager creates a brand new cri manager.
func NewCriManager(ctrMgr ContainerMgr, imgMgr ImageMgr) (*CriManager, error) {
	c := &CriManager{
		ContainerMgr: ctrMgr,
		ImageMgr:     imgMgr,
	}
	return c, nil
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
func (c *CriManager) RunPodSandbox(ctx context.Context, r *runtime.RunPodSandboxRequest) (*runtime.RunPodSandboxResponse, error) {
	config := r.GetConfig()

	// Step 1: Prepare image for the sandbox.
	// TODO: make sandbox image configurable.
	image := defaultSandboxImage

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

	// Step 3: Start the sandbox container.
	err = c.ContainerMgr.Start(ctx, createResp.ID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to start sandbox container for pod %q: %v", config.Metadata.Name, err)
	}

	// TODO: setup networking for the sandbox.

	return &runtime.RunPodSandboxResponse{PodSandboxId: createResp.ID}, nil
}

// StopPodSandbox stops the sandbox. If there are any running containers in the
// sandbox, they should be forcibly terminated.
func (c *CriManager) StopPodSandbox(ctx context.Context, r *runtime.StopPodSandboxRequest) (*runtime.StopPodSandboxResponse, error) {
	podSandboxID := r.GetPodSandboxId()

	opts := &ContainerListOption{All: true}
	filter := func(c *ContainerMeta) bool {
		return c.Config.Labels[sandboxIDLabelKey] == podSandboxID
	}

	containers, err := c.ContainerMgr.List(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to stop sandbox %q: %v", podSandboxID, err)
	}

	// Stop all containers in the sandbox.
	for _, container := range containers {
		err = c.ContainerMgr.Stop(ctx, container.ID, defaultStopTimeout)
		if err != nil {
			// TODO: log an error message or break?
		}
	}

	// TODO: tear down sandbox's network.

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

	opts := &ContainerListOption{All: true}
	filter := func(c *ContainerMeta) bool {
		return c.Config.Labels[sandboxIDLabelKey] == podSandboxID
	}

	containers, err := c.ContainerMgr.List(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to remove sandbox %q: %v", podSandboxID, err)
	}

	// Remove all containers in the sandbox.
	for _, container := range containers {
		err = c.ContainerMgr.Remove(ctx, container.ID, &ContainerRemoveOption{Volume: true, Force: true})
		if err != nil {
			// TODO: log an error message or break?
		}
	}

	// Remove the sandbox container.
	err = c.ContainerMgr.Remove(ctx, podSandboxID, &ContainerRemoveOption{Volume: true, Force: true})
	if err != nil {
		return nil, fmt.Errorf("failed to remove sandbox %q: %v", podSandboxID, err)
	}

	return &runtime.RemovePodSandboxResponse{}, nil
}

// PodSandboxStatus returns the status of the PodSandbox.
func (c *CriManager) PodSandboxStatus(ctx context.Context, r *runtime.PodSandboxStatusRequest) (*runtime.PodSandboxStatusResponse, error) {
	podSandboxID := r.GetPodSandboxId()
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
	status := &runtime.PodSandboxStatus{
		Id:          sandbox.ID,
		State:       state,
		CreatedAt:   createdAt,
		Metadata:    metadata,
		Labels:      labels,
		Annotations: annotations,
		// TODO: network status and linux specific pod status.
	}

	return &runtime.PodSandboxStatusResponse{Status: status}, nil
}

// ListPodSandbox returns a list of Sandbox.
func (c *CriManager) ListPodSandbox(ctx context.Context, r *runtime.ListPodSandboxRequest) (*runtime.ListPodSandboxResponse, error) {
	opts := &ContainerListOption{All: true}
	filter := func(c *ContainerMeta) bool {
		return c.Config.Labels[containerTypeLabelKey] == containerTypeLabelSandbox
	}

	// Filter *only* (sandbox) containers.
	sandboxList, err := c.ContainerMgr.List(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list sandbox: %v", err)
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
	createConfig := &apitypes.ContainerCreateConfig{
		ContainerConfig: apitypes.ContainerConfig{
			// TODO: wait for them to be fully supported.
			// Entrypoint:		config.Command,
			// Cmd:			config.Args,
			Env:        generateEnvList(config.GetEnvs()),
			Image:      image,
			WorkingDir: config.WorkingDir,
			Labels:     labels,
			// Interactive containers:
			OpenStdin: config.Stdin,
			StdinOnce: config.StdinOnce,
			Tty:       config.Tty,
		},
		HostConfig: &apitypes.HostConfig{
		// TODO: generate mount bindings.
		},
		NetworkingConfig: &apitypes.NetworkingConfig{},
	}
	c.updateCreateConfig(createConfig, config, sandboxConfig, podSandboxID)

	// TODO: devices and security option configurations.

	containerName := makeContainerName(sandboxConfig, config)

	createResp, err := c.ContainerMgr.Create(ctx, containerName, createConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create container for sandbox %q: %v", podSandboxID, err)
	}

	return &runtime.CreateContainerResponse{ContainerId: createResp.ID}, nil
}

// StartContainer starts the container.
func (c *CriManager) StartContainer(ctx context.Context, r *runtime.StartContainerRequest) (*runtime.StartContainerResponse, error) {
	containerID := r.GetContainerId()

	err := c.ContainerMgr.Start(ctx, containerID, "")
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

	err := c.ContainerMgr.Remove(ctx, containerID, &ContainerRemoveOption{Volume: true, Force: true})
	if err != nil {
		return nil, fmt.Errorf("failed to remove container %q: %v", containerID, err)
	}

	return &runtime.RemoveContainerResponse{}, nil
}

// ListContainers lists all containers matching the filter.
func (c *CriManager) ListContainers(ctx context.Context, r *runtime.ListContainersRequest) (*runtime.ListContainersResponse, error) {
	opts := &ContainerListOption{All: true}
	filter := func(c *ContainerMeta) bool {
		return c.Config.Labels[containerTypeLabelKey] == containerTypeLabelContainer
	}

	// Filter *only* (non-sandbox) containers.
	containerList, err := c.ContainerMgr.List(ctx, filter, opts)
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
		if finishedAt != 0 {
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

	status := &runtime.ContainerStatus{
		Id:          container.ID,
		Metadata:    metadata,
		Image:       &runtime.ImageSpec{Image: container.Config.Image},
		ImageRef:    container.Image,
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
		// TODO: LogPath.
	}

	return &runtime.ContainerStatusResponse{Status: status}, nil
}

// ContainerStats returns stats of the container. If the container does not
// exist, the call returns an error.
func (c *CriManager) ContainerStats(ctx context.Context, r *runtime.ContainerStatsRequest) (*runtime.ContainerStatsResponse, error) {
	return nil, fmt.Errorf("ContainerStats Not Implemented Yet")
}

// ListContainerStats returns stats of all running containers.
func (c *CriManager) ListContainerStats(ctx context.Context, r *runtime.ListContainerStatsRequest) (*runtime.ListContainerStatsResponse, error) {
	return nil, fmt.Errorf("ListContainerStats Not Implemented Yet")
}

// UpdateContainerResources updates ContainerConfig of the container.
func (c *CriManager) UpdateContainerResources(ctx context.Context, r *runtime.UpdateContainerResourcesRequest) (*runtime.UpdateContainerResourcesResponse, error) {
	return nil, fmt.Errorf("UpdateContainerResources Not Implemented Yet")
}

// ExecSync executes a command in the container, and returns the stdout output.
// If command exits with a non-zero exit code, an error is returned.
func (c *CriManager) ExecSync(ctx context.Context, r *runtime.ExecSyncRequest) (*runtime.ExecSyncResponse, error) {
	return nil, fmt.Errorf("ExecSync Not Implemented Yet")
}

// Exec prepares a streaming endpoint to execute a command in the container, and returns the address.
func (c *CriManager) Exec(ctx context.Context, r *runtime.ExecRequest) (*runtime.ExecResponse, error) {
	return nil, fmt.Errorf("Exec Not Implemented Yet")
}

// Attach prepares a streaming endpoint to attach to a running container, and returns the address.
func (c *CriManager) Attach(ctx context.Context, r *runtime.AttachRequest) (*runtime.AttachResponse, error) {
	return nil, fmt.Errorf("Attach Not Implemented Yet")
}

// PortForward prepares a streaming endpoint to forward ports from a PodSandbox, and returns the address.
func (c *CriManager) PortForward(ctx context.Context, r *runtime.PortForwardRequest) (*runtime.PortForwardResponse, error) {
	return nil, fmt.Errorf("PortForward Not Implemented Yet")
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

	// TODO: check network status of CRI when it is ready.

	return &runtime.StatusResponse{
		Status: &runtime.RuntimeStatus{Conditions: []*runtime.RuntimeCondition{
			runtimeCondition,
			networkCondition,
		}},
	}, nil
}

// ListImages lists existing images.
func (c *CriManager) ListImages(ctx context.Context, r *runtime.ListImagesRequest) (*runtime.ListImagesResponse, error) {
	// TODO: handle image list filters.
	imageList, err := c.ImageMgr.ListImages(ctx, "")
	if err != nil {
		return nil, err
	}

	images := make([]*runtime.Image, 0, len(imageList))
	for _, i := range imageList {
		image, err := imageToCriImage(&i)
		if err != nil {
			// TODO: log an error message?
			continue
		}
		images = append(images, image)
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
		return nil, err
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
	ref, err := reference.Parse(imageRef)
	if err != nil {
		return nil, err
	}

	err = c.ImageMgr.PullImage(ctx, ref.Name, ref.Tag, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}

	imageInfo, err := c.ImageMgr.GetImage(ctx, ref.String())
	if err != nil {
		return nil, err
	}

	return &runtime.PullImageResponse{ImageRef: imageInfo.ID}, nil
}

// RemoveImage removes the image.
func (c *CriManager) RemoveImage(ctx context.Context, r *runtime.RemoveImageRequest) (*runtime.RemoveImageResponse, error) {
	imageRef := r.GetImage().GetImage()
	ref, err := reference.Parse(imageRef)
	if err != nil {
		return nil, err
	}

	imageInfo, err := c.ImageMgr.GetImage(ctx, ref.String())
	if err != nil {
		return nil, err
	}

	err = c.ImageMgr.RemoveImage(ctx, imageInfo, &ImageRemoveOption{})
	if err != nil {
		return nil, err
	}

	return &runtime.RemoveImageResponse{}, nil
}

// ImageFsInfo returns information of the filesystem that is used to store images.
func (c *CriManager) ImageFsInfo(ctx context.Context, r *runtime.ImageFsInfoRequest) (*runtime.ImageFsInfoResponse, error) {
	return nil, fmt.Errorf("ImageFsInfo Not Implemented Yet")
}
