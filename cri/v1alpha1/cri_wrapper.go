package v1alpha1

import (
	"github.com/alibaba/pouch/cri/stream"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
)

// CriWrapper wraps CriManager and logs each operation for debugging convenice.
type CriWrapper struct {
	*CriManager
}

// NewCriWrapper creates a brand new CriWrapper.
func NewCriWrapper(c *CriManager) *CriWrapper {
	return &CriWrapper{CriManager: c}
}

// StreamServerStart starts the stream server of CRI.
func (c *CriWrapper) StreamServerStart() (err error) {
	logrus.Infof("StreamServerStart starts stream server of cri manager")
	defer func() {
		if err != nil {
			logrus.Errorf("failed to start StreamServer: %v", err)
		} else {
			logrus.Infof("success to start StreamServer of cri manager")
		}
	}()
	return c.CriManager.StreamServerStart()
}

// StreamRouter returns the router of Stream Server.
func (c *CriWrapper) StreamRouter() stream.Router {
	return c.CriManager.StreamRouter()
}

// Version returns the runtime name, runtime version and runtime API version.
func (c *CriWrapper) Version(ctx context.Context, r *runtime.VersionRequest) (res *runtime.VersionResponse, err error) {
	logrus.Debugf("Version shows the basic information of cri Manager")
	defer func() {
		if err != nil {
			logrus.Errorf("failed to get version: %v", err)
		} else {
			logrus.Debugf("success to get version")
		}
	}()
	return c.CriManager.Version(ctx, r)
}

// RunPodSandbox creates and starts a pod-level sandbox. Runtimes should ensure
// the sandbox is in ready state.
func (c *CriWrapper) RunPodSandbox(ctx context.Context, r *runtime.RunPodSandboxRequest) (res *runtime.RunPodSandboxResponse, err error) {
	// NOTE: maybe we should log the verbose configuration in higher log level.
	logrus.Infof("RunPodSandbox with config %+v", r.GetConfig())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to run PodSandbox: %+v, %v", r.GetConfig().GetMetadata(), err)
		} else {
			logrus.Infof("success to run PodSandbox: %+v, return sandbox id: %q", r.GetConfig().GetMetadata(), res.GetPodSandboxId())
		}
	}()
	return c.CriManager.RunPodSandbox(ctx, r)
}

// StopPodSandbox stops the sandbox. If there are any running containers in the
// sandbox, they should be forcibly terminated.
func (c *CriWrapper) StopPodSandbox(ctx context.Context, r *runtime.StopPodSandboxRequest) (res *runtime.StopPodSandboxResponse, err error) {
	logrus.Infof("StopPodSandbox for %q", r.GetPodSandboxId())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to stop PodSandbox: %q, %v", r.GetPodSandboxId(), err)
		} else {
			logrus.Infof("success to stop PodSandbox: %q", r.GetPodSandboxId())
		}
	}()
	return c.CriManager.StopPodSandbox(ctx, r)
}

// RemovePodSandbox removes the sandbox. If there are running containers in the
// sandbox, they should be forcibly removed.
func (c *CriWrapper) RemovePodSandbox(ctx context.Context, r *runtime.RemovePodSandboxRequest) (res *runtime.RemovePodSandboxResponse, err error) {
	logrus.Infof("RemovePodSandbox for %q", r.GetPodSandboxId())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to remove PodSandbox: %q, %v", r.GetPodSandboxId(), err)
		} else {
			logrus.Infof("success to remove PodSandbox: %q", r.GetPodSandboxId())
		}
	}()
	return c.CriManager.RemovePodSandbox(ctx, r)
}

// PodSandboxStatus returns the status of the PodSandbox.
func (c *CriWrapper) PodSandboxStatus(ctx context.Context, r *runtime.PodSandboxStatusRequest) (res *runtime.PodSandboxStatusResponse, err error) {
	logrus.Debugf("PodSandboxStatus for %q", r.GetPodSandboxId())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to get PodSandboxStatus: %q, %v", r.GetPodSandboxId(), err)
		} else {
			logrus.Debugf("success to get PodSandboxStatus: %q, %+v", r.GetPodSandboxId(), res.GetStatus())
		}
	}()
	return c.CriManager.PodSandboxStatus(ctx, r)
}

// ListPodSandbox returns a list of Sandbox.
func (c *CriWrapper) ListPodSandbox(ctx context.Context, r *runtime.ListPodSandboxRequest) (res *runtime.ListPodSandboxResponse, err error) {
	logrus.Debugf("ListPodSandbox with filter %+v", r.GetFilter())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to ListPodSandbox: %v", err)
		} else {
			// NOTE: maybe log detailed sandbox items with higher log level.
			logrus.Debugf("success to ListPodSandbox: %+v", res.Items)
		}
	}()
	return c.CriManager.ListPodSandbox(ctx, r)
}

// CreateContainer creates a new container in the given PodSandbox.
func (c *CriWrapper) CreateContainer(ctx context.Context, r *runtime.CreateContainerRequest) (res *runtime.CreateContainerResponse, err error) {
	// NOTE: maybe we should log the verbose configuration in higher log level.
	logrus.Infof("CreateContainer within sandbox %q with container config %+v and sandbox %+v",
		r.GetPodSandboxId(), r.GetConfig(), r.GetSandboxConfig())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to create container within sandbox %q with container config %+v: %v",
				r.GetPodSandboxId(), r.GetConfig().GetMetadata(), err)
		} else {
			logrus.Infof("success to create container within sandbox %q with container config %+v, return container id %q",
				r.GetPodSandboxId(), r.GetConfig().GetMetadata(), res.GetContainerId())
		}
	}()
	return c.CriManager.CreateContainer(ctx, r)
}

// StartContainer starts the container.
func (c *CriWrapper) StartContainer(ctx context.Context, r *runtime.StartContainerRequest) (res *runtime.StartContainerResponse, err error) {
	logrus.Infof("StartContainer for %q", r.GetContainerId())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to start container: %q, %v", r.GetContainerId(), err)
		} else {
			logrus.Infof("success to start container: %q", r.GetContainerId())
		}
	}()
	return c.CriManager.StartContainer(ctx, r)
}

// StopContainer stops a running container with a grace period (i.e., timeout).
func (c *CriWrapper) StopContainer(ctx context.Context, r *runtime.StopContainerRequest) (res *runtime.StopContainerResponse, err error) {
	logrus.Infof("StopContainer for %q with timeout %d (s)", r.GetContainerId(), r.GetTimeout())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to stop container: %q, %v", r.GetContainerId(), err)
		} else {
			logrus.Infof("success to stop container: %q", r.GetContainerId())
		}
	}()
	return c.CriManager.StopContainer(ctx, r)
}

// RemoveContainer removes the container.
func (c *CriWrapper) RemoveContainer(ctx context.Context, r *runtime.RemoveContainerRequest) (res *runtime.RemoveContainerResponse, err error) {
	logrus.Infof("RemoveContainer for %q", r.GetContainerId())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to remove container: %q, %v", r.GetContainerId(), err)
		} else {
			logrus.Infof("success to remove container: %q", r.GetContainerId())
		}
	}()
	return c.CriManager.RemoveContainer(ctx, r)
}

// ListContainers lists all containers matching the filter.
func (c *CriWrapper) ListContainers(ctx context.Context, r *runtime.ListContainersRequest) (res *runtime.ListContainersResponse, err error) {
	logrus.Debugf("ListContainers with filter %+v", r.GetFilter())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to list containers with filter %+v: %v", r.GetFilter(), err)
		} else {
			// NOTE: maybe log detailed container items with higher log level.
			logrus.Debugf("success to list containers with filter: %+v", r.GetFilter())
		}
	}()
	return c.CriManager.ListContainers(ctx, r)
}

// ContainerStatus inspects the container and returns the status.
func (c *CriWrapper) ContainerStatus(ctx context.Context, r *runtime.ContainerStatusRequest) (res *runtime.ContainerStatusResponse, err error) {
	logrus.Debugf("ContainerStatus for %q", r.GetContainerId())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to get ContainerStatus: %q, %v", r.GetContainerId(), err)
		} else {
			logrus.Debugf("success to get ContainerStatus: %q, %+v", r.GetContainerId(), res.GetStatus())
		}
	}()
	return c.CriManager.ContainerStatus(ctx, r)
}

// ContainerStats returns stats of the container. If the container does not
// exist, the call returns an error.
func (c *CriWrapper) ContainerStats(ctx context.Context, r *runtime.ContainerStatsRequest) (res *runtime.ContainerStatsResponse, err error) {
	logrus.Debugf("ContainerStats for %q", r.GetContainerId())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to get ContainerStats: %q, %v", r.GetContainerId(), err)
		} else {
			logrus.Debugf("success to get ContainerStats: %q, %+v", r.GetContainerId(), res.GetStats())
		}
	}()
	return c.CriManager.ContainerStats(ctx, r)
}

// ListContainerStats returns stats of all running containers.
func (c *CriWrapper) ListContainerStats(ctx context.Context, r *runtime.ListContainerStatsRequest) (res *runtime.ListContainerStatsResponse, err error) {
	logrus.Debugf("ListContainerStats with filter %+v", r.GetFilter())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to get ListContainerStats: %v", err)
		} else {
			logrus.Debugf("success to get ListContainerStats: %+v", res.GetStats())
		}
	}()
	return c.CriManager.ListContainerStats(ctx, r)
}

// UpdateContainerResources updates ContainerConfig of the container.
func (c *CriWrapper) UpdateContainerResources(ctx context.Context, r *runtime.UpdateContainerResourcesRequest) (res *runtime.UpdateContainerResourcesResponse, err error) {
	logrus.Infof("UpdateContainerResources for %q with %+v", r.GetContainerId(), r.GetLinux())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to update ContainerResources: %q", r.GetContainerId())
		} else {
			logrus.Infof("success to update ContainerResources: %q", r.GetContainerId())
		}
	}()
	return c.CriManager.UpdateContainerResources(ctx, r)
}

// ExecSync executes a command in the container, and returns the stdout output.
// If command exits with a non-zero exit code, an error is returned.
func (c *CriWrapper) ExecSync(ctx context.Context, r *runtime.ExecSyncRequest) (res *runtime.ExecSyncResponse, err error) {
	logrus.Infof("ExecSync for %q with command %+v and timeout %d (s)", r.GetContainerId(), r.GetCmd(), r.GetTimeout())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to ExecSync: %q, %v", r.GetContainerId(), err)
		} else {
			logrus.Infof("success to ExecSync %q, return with exit code %d", r.GetContainerId(), res.GetExitCode())
		}
	}()
	return c.CriManager.ExecSync(ctx, r)
}

// Exec prepares a streaming endpoint to execute a command in the container, and returns the address.
func (c *CriWrapper) Exec(ctx context.Context, r *runtime.ExecRequest) (res *runtime.ExecResponse, err error) {
	logrus.Infof("Exec for %q with command %+v, tty %v and stdin %v",
		r.GetContainerId(), r.GetCmd(), r.GetTty(), r.GetStdin())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to Exec %q, %v", r.GetContainerId(), err)
		} else {
			logrus.Infof("success to Exec %q, return URL %q", r.GetContainerId(), res.GetUrl())
		}
	}()
	return c.CriManager.Exec(ctx, r)
}

// Attach prepares a streaming endpoint to attach to a running container, and returns the address.
func (c *CriWrapper) Attach(ctx context.Context, r *runtime.AttachRequest) (res *runtime.AttachResponse, err error) {
	logrus.Infof("Attach for %q with tty %v and stdin %v", r.GetContainerId(), r.GetTty(), r.GetStdin())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to attach: %q, %v", r.GetContainerId(), err)
		} else {
			logrus.Infof("success to attach: %q, return URL %q", r.GetContainerId(), res.GetUrl())
		}
	}()
	return c.CriManager.Attach(ctx, r)
}

// PortForward prepares a streaming endpoint to forward ports from a PodSandbox, and returns the address.
func (c *CriWrapper) PortForward(ctx context.Context, r *runtime.PortForwardRequest) (res *runtime.PortForwardResponse, err error) {
	logrus.Infof("Portforward for %q port %v", r.GetPodSandboxId(), r.GetPort())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to portforward: %q, %v", r.GetPodSandboxId(), err)
		} else {
			logrus.Infof("success to portforward: %q, return URL %q", r.GetPodSandboxId(), res.GetUrl())
		}
	}()
	return c.CriManager.PortForward(ctx, r)
}

// UpdateRuntimeConfig updates the runtime config. Currently only handles podCIDR updates.
func (c *CriWrapper) UpdateRuntimeConfig(ctx context.Context, r *runtime.UpdateRuntimeConfigRequest) (res *runtime.UpdateRuntimeConfigResponse, err error) {
	logrus.Infof("UpdateRuntimeConfig with config %+v", r.GetRuntimeConfig())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to update RuntimeConfig: %v", err)
		} else {
			logrus.Infof("success to update RuntimeConfig")
		}
	}()
	return c.CriManager.UpdateRuntimeConfig(ctx, r)
}

// Status returns the status of the runtime.
func (c *CriWrapper) Status(ctx context.Context, r *runtime.StatusRequest) (res *runtime.StatusResponse, err error) {
	logrus.Debugf("Status of cri manager")
	defer func() {
		if err != nil {
			logrus.Errorf("failed to get status: %v", err)
		} else {
			logrus.Debugf("success to get status: %+v", res.GetStatus())
		}
	}()
	return c.CriManager.Status(ctx, r)
}

// ListImages lists existing images.
func (c *CriWrapper) ListImages(ctx context.Context, r *runtime.ListImagesRequest) (res *runtime.ListImagesResponse, err error) {
	logrus.Debugf("ListImages with filter %+v", r.GetFilter())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to list images with filter %+v: %v", r.GetFilter(), err)
		} else {
			// NOTE: maybe log detailed image items with higher log level.
			logrus.Debugf("success to list images with filter: %+v", r.GetFilter())
		}
	}()
	return c.CriManager.ListImages(ctx, r)
}

// ImageStatus returns the status of the image, returns nil if the image isn't present.
func (c *CriWrapper) ImageStatus(ctx context.Context, r *runtime.ImageStatusRequest) (res *runtime.ImageStatusResponse, err error) {
	logrus.Debugf("ImageStatus for %q", r.GetImage().GetImage())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to get ImageStatus: %q, %v", r.GetImage().GetImage(), err)
		} else {
			logrus.Debugf("success to get ImageStatus: %q, %+v",
				r.GetImage().GetImage(), res.GetImage())
		}
	}()
	return c.CriManager.ImageStatus(ctx, r)
}

// PullImage pulls an image with authentication config.
func (c *CriWrapper) PullImage(ctx context.Context, r *runtime.PullImageRequest) (res *runtime.PullImageResponse, err error) {
	logrus.Infof("PullImage %q with auth config %+v", r.GetImage().GetImage(), r.GetAuth())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to pull image: %q, %v", r.GetImage().GetImage(), err)
		} else {
			logrus.Infof("success to pull image %q, return image reference %q",
				r.GetImage().GetImage(), res.GetImageRef())
		}
	}()
	return c.CriManager.PullImage(ctx, r)
}

// RemoveImage removes the image.
func (c *CriWrapper) RemoveImage(ctx context.Context, r *runtime.RemoveImageRequest) (res *runtime.RemoveImageResponse, err error) {
	logrus.Infof("RemoveImage %q", r.GetImage().GetImage())
	defer func() {
		if err != nil {
			logrus.Errorf("failed to remove image: %q, %v", r.GetImage().GetImage(), err)
		} else {
			logrus.Infof("success to remove image: %q", r.GetImage().GetImage())
		}
	}()
	return c.CriManager.RemoveImage(ctx, r)
}

// ImageFsInfo returns information of the filesystem that is used to store images.
func (c *CriWrapper) ImageFsInfo(ctx context.Context, r *runtime.ImageFsInfoRequest) (res *runtime.ImageFsInfoResponse, err error) {
	logrus.Debugf("ImageFsInfo of cri manager")
	defer func() {
		if err != nil {
			logrus.Errorf("faild to get ImageFsInfo: %v", err)
		} else {
			logrus.Debugf("success to get ImageFsInfo, return filesystem info %+v", res.GetImageFilesystems())
		}
	}()
	return c.CriManager.ImageFsInfo(ctx, r)
}
