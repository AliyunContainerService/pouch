package mgr

import (
	"fmt"

	"golang.org/x/net/context"
	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
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
	return nil, fmt.Errorf("Version Not Implemented Yet")
}

// RunPodSandbox creates and starts a pod-level sandbox. Runtimes should ensure
// the sandbox is in ready state.
func (c *CriManager) RunPodSandbox(ctx context.Context, r *runtime.RunPodSandboxRequest) (*runtime.RunPodSandboxResponse, error) {
	return nil, fmt.Errorf("RunPodSandbox Not Implemented Yet")
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

// CreateContainer creates a new container in the given PodSandbox.
func (c *CriManager) CreateContainer(ctx context.Context, r *runtime.CreateContainerRequest) (*runtime.CreateContainerResponse, error) {
	return nil, fmt.Errorf("CreateContainer Not Implemented Yet")
}

// StartContainer starts the container.
func (c *CriManager) StartContainer(ctx context.Context, r *runtime.StartContainerRequest) (*runtime.StartContainerResponse, error) {
	return nil, fmt.Errorf("StartContainer Not Implemented Yet")
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
	return nil, fmt.Errorf("Status Not Implemented Yet")
}

// ListImages lists existing images.
func (c *CriManager) ListImages(ctx context.Context, r *runtime.ListImagesRequest) (*runtime.ListImagesResponse, error) {
	return nil, fmt.Errorf("ListImages Not Implemented Yet")
}

// ImageStatus returns the status of the image, returns nil if the image isn't present.
func (c *CriManager) ImageStatus(ctx context.Context, r *runtime.ImageStatusRequest) (*runtime.ImageStatusResponse, error) {
	return nil, fmt.Errorf("ImageStatus Not Implemented Yet")
}

// PullImage pulls an image with authentication config.
func (c *CriManager) PullImage(ctx context.Context, r *runtime.PullImageRequest) (*runtime.PullImageResponse, error) {
	return nil, fmt.Errorf("PullImage Not Implemented Yet")
}

// RemoveImage removes the image.
func (c *CriManager) RemoveImage(ctx context.Context, r *runtime.RemoveImageRequest) (*runtime.RemoveImageResponse, error) {
	return nil, fmt.Errorf("RemoveImage Not Implemented Yet")
}

// ImageFsInfo returns information of the filesystem that is used to store images.
func (c *CriManager) ImageFsInfo(ctx context.Context, r *runtime.ImageFsInfoRequest) (*runtime.ImageFsInfoResponse, error) {
	return nil, fmt.Errorf("ImageFsInfo Not Implemented Yet")
}
