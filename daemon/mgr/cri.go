package mgr

import (
	"bytes"
	"fmt"
	"strings"

	apitypes "github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/config"
	"github.com/alibaba/pouch/pkg/reference"
	"github.com/alibaba/pouch/version"

	// NOTE: "golang.org/x/net/context" is compatible with standard "context" in golang1.7+.
	"github.com/go-openapi/strfmt"
	"golang.org/x/net/context"
	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
	"k8s.io/kubernetes/pkg/kubelet/server/streaming"
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

// CriMgr as an interface defines all operations against CRI.
type CriMgr interface {
	// RuntimeServiceServer is interface of CRI runtime service.
	runtime.RuntimeServiceServer

	// ImageServiceServer is interface of CRI image service.
	runtime.ImageServiceServer

	// StreamServerStart starts stream server of CRI.
	StreamServerStart() error
}

// CriManager is an implementation of interface CriMgr.
type CriManager struct {
	ContainerMgr ContainerMgr
	ImageMgr     ImageMgr
	// StreamServer is the stream server of CRI serves container streaming request.
	StreamServer streaming.Server
}

// NewCriManager creates a brand new cri manager.
func NewCriManager(cfg *config.Config, ctrMgr ContainerMgr, imgMgr ImageMgr) (*CriManager, error) {
	c := &CriManager{
		ContainerMgr: ctrMgr,
		ImageMgr:     imgMgr,
	}

	var err error
	c.StreamServer, err = newStreamServer(ctrMgr, cfg.StreamServerAddress, cfg.StreamServerPort)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream server: %v", err)
	}
	return c, nil
}

// StreamServerStart starts stream server of CRI.
func (c *CriManager) StreamServerStart() error {
	return c.StreamServer.Start(true)
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

func makeLabels(labels, annotations map[string]string) map[string]string {
	m := make(map[string]string)

	for k, v := range labels {
		m[k] = v
	}
	for k, v := range annotations {
		// Use prefix to distinguish between annotations and labels.
		m[fmt.Sprintf("%s%s", annotationPrefix, k)] = v
	}

	return m
}

// makeSandboxPouchConfig returns apitypes.ContainerCreateConfig based on runtimeapi.PodSandboxConfig.
func (c *CriManager) makeSandboxPouchConfig(config *runtime.PodSandboxConfig, image string) (*apitypes.ContainerCreateConfig, error) {
	// Merge annotations and labels because pouch supports only labels.
	labels := makeLabels(config.GetLabels(), config.GetAnnotations())
	// Apply a label to distinguish sandboxes from regular containers.
	labels[containerTypeLabelKey] = containerTypeLabelSandbox

	hc := &apitypes.HostConfig{}
	createConfig := &apitypes.ContainerCreateConfig{
		ContainerConfig: apitypes.ContainerConfig{
			Hostname: strfmt.Hostname(config.Hostname),
			Image:    image,
			Labels:   labels,
		},
		HostConfig:       hc,
		NetworkingConfig: &apitypes.NetworkingConfig{},
	}

	return createConfig, nil
}

// makeSandboxName generates sandbox name from sandbox metadata. The name
// generated is unique as long as sandbox metadata is unique.
func makeSandboxName(c *runtime.PodSandboxConfig) string {
	return strings.Join([]string{
		kubePrefix,                            // 0
		sandboxContainerName,                  // 1
		c.Metadata.Name,                       // 2
		c.Metadata.Namespace,                  // 3
		c.Metadata.Uid,                        // 4
		fmt.Sprintf("%d", c.Metadata.Attempt), // 5
	}, nameDelimiter)
}

// RunPodSandbox creates and starts a pod-level sandbox. Runtimes should ensure
// the sandbox is in ready state.
func (c *CriManager) RunPodSandbox(ctx context.Context, r *runtime.RunPodSandboxRequest) (*runtime.RunPodSandboxResponse, error) {
	config := r.GetConfig()

	// Step 1: Prepare image for the sandbox.
	// TODO: make sandbox image configurable.
	image := defaultSandboxImage

	// TODO: make sure the image exists.

	// Step 2: Create the sandbox container.
	createConfig, err := c.makeSandboxPouchConfig(config, image)
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
	return nil, fmt.Errorf("StopPodSandbox Not Implemented Yet")
}

// RemovePodSandbox removes the sandbox. If there are running containers in the
// sandbox, they should be forcibly removed.
func (c *CriManager) RemovePodSandbox(ctx context.Context, r *runtime.RemovePodSandboxRequest) (*runtime.RemovePodSandboxResponse, error) {
	return nil, fmt.Errorf("RemovePodSandbox Not Implemented Yet")
}

// PodSandboxStatus returns the status of the PodSandbox.
func (c *CriManager) PodSandboxStatus(ctx context.Context, r *runtime.PodSandboxStatusRequest) (*runtime.PodSandboxStatusResponse, error) {
	return nil, fmt.Errorf("PodSandboxStatus Not Implemented Yet")
}

// ListPodSandbox returns a list of Sandbox.
func (c *CriManager) ListPodSandbox(ctx context.Context, r *runtime.ListPodSandboxRequest) (*runtime.ListPodSandboxResponse, error) {
	return nil, fmt.Errorf("ListPodSandbox Not Implemented Yet")
}

// generateEnvList converts KeyValue list to a list of strings, in the form of
// '<key>=<value>', which can be understood by pouch.
func generateEnvList(envs []*runtime.KeyValue) (result []string) {
	for _, env := range envs {
		result = append(result, fmt.Sprintf("%s=%s", env.Key, env.Value))
	}
	return
}

// modifyContainerNamespaceOptions apply namespace options for container.
func modifyContainerNamespaceOptions(nsOpts *runtime.NamespaceOption, podSandboxID string, hostConfig *apitypes.HostConfig) {
	sandboxNSMode := fmt.Sprintf("container:%v", podSandboxID)

	hostConfig.PidMode = sandboxNSMode
	hostConfig.NetworkMode = sandboxNSMode
	hostConfig.IpcMode = sandboxNSMode
	hostConfig.UTSMode = sandboxNSMode
}

// applyContainerSecurityContext updates pouch container options according to security context.
func applyContainerSecurityContext(lc *runtime.LinuxContainerConfig, podSandboxID string, config *apitypes.ContainerConfig, hc *apitypes.HostConfig) error {
	// TODO: modify Config and HostConfig.

	modifyContainerNamespaceOptions(lc.SecurityContext.GetNamespaceOptions(), podSandboxID, hc)

	return nil
}

// Apply Linux-specific options if applicable.
func (c *CriManager) updateCreateConfig(createConfig *apitypes.ContainerCreateConfig, config *runtime.ContainerConfig, sandboxConfig *runtime.PodSandboxConfig, podSandboxID string) error {
	if lc := config.GetLinux(); lc != nil {
		// TODO: resource restriction.

		// Apply security context.
		if err := applyContainerSecurityContext(lc, podSandboxID, &createConfig.ContainerConfig, createConfig.HostConfig); err != nil {
			return fmt.Errorf("failed to apply container security context for container %q: %v", config.Metadata.Name, err)
		}
	}

	// TODO: apply cgroupParent derived from the sandbox config.

	return nil
}

func makeContainerName(s *runtime.PodSandboxConfig, c *runtime.ContainerConfig) string {
	return strings.Join([]string{
		kubePrefix,                            // 0
		c.Metadata.Name,                       // 1
		s.Metadata.Name,                       // 2: sandbox name
		s.Metadata.Namespace,                  // 3: sandbox namespace
		s.Metadata.Uid,                        // 4: sandbox uid
		fmt.Sprintf("%d", c.Metadata.Attempt), // 5
	}, nameDelimiter)
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
	return nil, fmt.Errorf("StopContainer Not Implemented Yet")
}

// RemoveContainer removes the container.
func (c *CriManager) RemoveContainer(ctx context.Context, r *runtime.RemoveContainerRequest) (*runtime.RemoveContainerResponse, error) {
	return nil, fmt.Errorf("RemoveContainer Not Implemented Yet")
}

// ListContainers lists all containers matching the filter.
func (c *CriManager) ListContainers(ctx context.Context, r *runtime.ListContainersRequest) (*runtime.ListContainersResponse, error) {
	return nil, fmt.Errorf("ListContainers Not Implemented Yet")
}

// ContainerStatus inspects the container and returns the status.
func (c *CriManager) ContainerStatus(ctx context.Context, r *runtime.ContainerStatusRequest) (*runtime.ContainerStatusResponse, error) {
	return nil, fmt.Errorf("ContainerStatus Not Implemented Yet")
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
	// TODO: check that container is running.
	return c.StreamServer.GetExec(r)
}

// Attach prepares a streaming endpoint to attach to a running container, and returns the address.
func (c *CriManager) Attach(ctx context.Context, r *runtime.AttachRequest) (*runtime.AttachResponse, error) {
	// TODO: check that container is running.
	return c.StreamServer.GetAttach(r)
}

// PortForward prepares a streaming endpoint to forward ports from a PodSandbox, and returns the address.
func (c *CriManager) PortForward(ctx context.Context, r *runtime.PortForwardRequest) (*runtime.PortForwardResponse, error) {
	// TODO: check that sandbox is running.
	return c.StreamServer.GetPortForward(r)
}

// UpdateRuntimeConfig updates the runtime config. Currently only handles podCIDR updates.
func (c *CriManager) UpdateRuntimeConfig(ctx context.Context, r *runtime.UpdateRuntimeConfigRequest) (*runtime.UpdateRuntimeConfigResponse, error) {
	return nil, fmt.Errorf("UpdateRuntimeConfig Not Implemented Yet")
}

// Status returns the status of the runtime.
func (c *CriManager) Status(ctx context.Context, r *runtime.StatusRequest) (*runtime.StatusResponse, error) {
	return nil, fmt.Errorf("Status Not Implemented Yet")
}

// imageToCriImage converts pouch image API to CRI image API.
func imageToCriImage(image *apitypes.ImageInfo) (*runtime.Image, error) {
	ref, err := reference.Parse(image.Name)
	if err != nil {
		return nil, err
	}

	size := uint64(image.Size)
	// TODO: improve type ImageInfo to include RepoTags and RepoDigests.
	return &runtime.Image{
		Id:          image.Digest,
		RepoTags:    []string{fmt.Sprintf("%s:%s", ref.Name, ref.Tag)},
		RepoDigests: []string{fmt.Sprintf("%s@%s", ref.Name, image.Digest)},
		Size_:       size,
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
	imageInfo, err := c.ImageMgr.GetImage(ctx, imageRef)
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

	imageInfo, err := c.ImageMgr.GetImage(ctx, imageRef)
	if err != nil {
		return nil, err
	}

	return &runtime.PullImageResponse{ImageRef: imageInfo.ID}, nil
}

// RemoveImage removes the image.
func (c *CriManager) RemoveImage(ctx context.Context, r *runtime.RemoveImageRequest) (*runtime.RemoveImageResponse, error) {
	imageRef := r.GetImage().GetImage()
	imageInfo, err := c.ImageMgr.GetImage(ctx, imageRef)
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
