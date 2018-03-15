package mgr

import (
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
			logrus.Errorf("StreamServerStart failed: %v", err)
		} else {
			logrus.Infof("StreamServerStart has started stream server of cri manager")
		}
	}()
	return c.CriManager.StreamServerStart()
}

// Version returns the runtime name, runtime version and runtime API version.
func (c *CriWrapper) Version(ctx context.Context, r *runtime.VersionRequest) (res *runtime.VersionResponse, err error) {
	logrus.Infof("Version shows the basic information of cri Manager")
	defer func() {
		if err != nil {
			logrus.Errorf("Version failed: %v", err)
		} else {
			logrus.Infof("Version has operated successfully")
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
			logrus.Errorf("RunPodSandbox for %+v failed, error: %v", r.GetConfig().GetMetadata(), err)
		} else {
			logrus.Infof("RunPodSandbox for %+v returns sandbox id %q", r.GetConfig().GetMetadata(), res.GetPodSandboxId())
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
			logrus.Errorf("StopPodSandbox for %q failed, error: %v", r.GetPodSandboxId(), err)
		} else {
			logrus.Infof("StopPodSandbox for %q returns successfully", r.GetPodSandboxId())
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
			logrus.Errorf("RemovePodSandbox for %q failed, error: %v", r.GetPodSandboxId(), err)
		} else {
			logrus.Infof("RemovePodSandbox %q returns successfully", r.GetPodSandboxId())
		}
	}()
	return c.CriManager.RemovePodSandbox(ctx, r)
}

// PodSandboxStatus returns the status of the PodSandbox.
func (c *CriWrapper) PodSandboxStatus(ctx context.Context, r *runtime.PodSandboxStatusRequest) (res *runtime.PodSandboxStatusResponse, err error) {
	logrus.Infof("PodSandboxStatus for %q", r.GetPodSandboxId())
	defer func() {
		if err != nil {
			logrus.Errorf("PodSandboxStatus for %q failed, error: %v", r.GetPodSandboxId(), err)
		} else {
			logrus.Infof("PodSandboxStatus for %q returns status %+v", r.GetPodSandboxId(), res.GetStatus())
		}
	}()
	return c.CriManager.PodSandboxStatus(ctx, r)
}

// ListPodSandbox returns a list of Sandbox.
func (c *CriWrapper) ListPodSandbox(ctx context.Context, r *runtime.ListPodSandboxRequest) (res *runtime.ListPodSandboxResponse, err error) {
	logrus.Infof("ListPodSandbox with filter %+v", r.GetFilter())
	defer func() {
		if err != nil {
			logrus.Errorf("ListPodSandbox failed, error: %v", err)
		} else {
			// NOTE: maybe log detailed sandbox items with higher log level.
			logrus.Infof("ListPodSandbox returns sandboxes list successfully")
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
			logrus.Errorf("CreateContainer within sandbox %q for %+v failed, error: %v",
				r.GetPodSandboxId(), r.GetConfig().GetMetadata(), err)
		} else {
			logrus.Infof("CreateContainer within sandbox %q for %+v returns container id %q",
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
			logrus.Errorf("StartContainer for %q failed, error: %v", r.GetContainerId(), err)
		} else {
			logrus.Infof("StartContainer for %q returns successfully", r.GetContainerId())
		}
	}()
	return c.CriManager.StartContainer(ctx, r)
}

// StopContainer stops a running container with a grace period (i.e., timeout).
func (c *CriWrapper) StopContainer(ctx context.Context, r *runtime.StopContainerRequest) (res *runtime.StopContainerResponse, err error) {
	logrus.Infof("StopContainer for %q with timeout %d (s)", r.GetContainerId(), r.GetTimeout())
	defer func() {
		if err != nil {
			logrus.Errorf("StopContainer for %q failed, error: %v", r.GetContainerId(), err)
		} else {
			logrus.Infof("StopContainer for %q returns successfully", r.GetContainerId())
		}
	}()
	return c.CriManager.StopContainer(ctx, r)
}

// RemoveContainer removes the container.
func (c *CriWrapper) RemoveContainer(ctx context.Context, r *runtime.RemoveContainerRequest) (res *runtime.RemoveContainerResponse, err error) {
	logrus.Infof("RemoveContainer for %q", r.GetContainerId())
	defer func() {
		if err != nil {
			logrus.Errorf("RemoveContainer for %q failed, error: %v", r.GetContainerId(), err)
		} else {
			logrus.Infof("RemoveContainer for %q returns successfully", r.GetContainerId())
		}
	}()
	return c.CriManager.RemoveContainer(ctx, r)
}

// ListContainers lists all containers matching the filter.
func (c *CriWrapper) ListContainers(ctx context.Context, r *runtime.ListContainersRequest) (res *runtime.ListContainersResponse, err error) {
	logrus.Infof("ListContainers with filter %+v", r.GetFilter())
	defer func() {
		if err != nil {
			logrus.Errorf("ListContainers with filter %+v failed, error: %v", r.GetFilter(), err)
		} else {
			// NOTE: maybe log detailed container items with higher log level.
			logrus.Infof("ListContainers with filter %+v returns container list successfully", r.GetFilter())
		}
	}()
	return c.CriManager.ListContainers(ctx, r)
}

// ContainerStatus inspects the container and returns the status.
func (c *CriWrapper) ContainerStatus(ctx context.Context, r *runtime.ContainerStatusRequest) (res *runtime.ContainerStatusResponse, err error) {
	logrus.Infof("ContainerStatus for %q", r.GetContainerId())
	defer func() {
		if err != nil {
			logrus.Errorf("ContainerStatus for %q failed, error: %v", r.GetContainerId(), err)
		} else {
			logrus.Infof("ContainerStatus for %q returns status %+v", r.GetContainerId(), res.GetStatus())
		}
	}()
	return c.CriManager.ContainerStatus(ctx, r)
}

// ContainerStats returns stats of the container. If the container does not
// exist, the call returns an error.
func (c *CriWrapper) ContainerStats(ctx context.Context, r *runtime.ContainerStatsRequest) (res *runtime.ContainerStatsResponse, err error) {
	logrus.Infof("ContainerStats for %q", r.GetContainerId())
	defer func() {
		if err != nil {
			logrus.Errorf("ContainerStats for %q failed, error: %v", r.GetContainerId(), err)
		} else {
			logrus.Infof("ContainerStats for %q returns stats %+v", r.GetContainerId(), res.GetStats())
		}
	}()
	return c.CriManager.ContainerStats(ctx, r)
}

// ListContainerStats returns stats of all running containers.
func (c *CriWrapper) ListContainerStats(ctx context.Context, r *runtime.ListContainerStatsRequest) (res *runtime.ListContainerStatsResponse, err error) {
	logrus.Infof("ListContainerStats with filter %+v", r.GetFilter())
	defer func() {
		if err != nil {
			logrus.Errorf("ListContainerStats failed, error: %v", err)
		} else {
			logrus.Infof("ListContainerStats returns stats %+v", res.GetStats())
		}
	}()
	return c.CriManager.ListContainerStats(ctx, r)
}

// UpdateContainerResources updates ContainerConfig of the container.
func (c *CriWrapper) UpdateContainerResources(ctx context.Context, r *runtime.UpdateContainerResourcesRequest) (res *runtime.UpdateContainerResourcesResponse, err error) {
	logrus.Infof("UpdateContainerResources for %q with %+v", r.GetContainerId(), r.GetLinux())
	defer func() {
		if err != nil {
			logrus.Errorf("UpdateContainerResources for %q failed", r.GetContainerId())
		} else {
			logrus.Infof("UpdateContainerResources for %q returns successfully", r.GetContainerId())
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
			logrus.Errorf("ExecSync for %q failed, error: %v", r.GetContainerId(), err)
		} else {
			logrus.Infof("ExecSync for %q returns with exit code %d", r.GetContainerId(), res.GetExitCode())
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
			logrus.Errorf("Exec for %q failed, error: %v", r.GetContainerId(), err)
		} else {
			logrus.Infof("Exec for %q returns URL %q", r.GetContainerId(), res.GetUrl())
		}
	}()
	return c.CriManager.Exec(ctx, r)
}

// Attach prepares a streaming endpoint to attach to a running container, and returns the address.
func (c *CriWrapper) Attach(ctx context.Context, r *runtime.AttachRequest) (res *runtime.AttachResponse, err error) {
	logrus.Infof("Attach for %q with tty %v and stdin %v", r.GetContainerId(), r.GetTty(), r.GetStdin())
	defer func() {
		if err != nil {
			logrus.Errorf("Attach for %q failed, error: %v", r.GetContainerId(), err)
		} else {
			logrus.Infof("Attach for %q returns URL %q", r.GetContainerId(), res.GetUrl())
		}
	}()
	return c.CriManager.Attach(ctx, r)
}

// PortForward prepares a streaming endpoint to forward ports from a PodSandbox, and returns the address.
func (c *CriWrapper) PortForward(ctx context.Context, r *runtime.PortForwardRequest) (res *runtime.PortForwardResponse, err error) {
	logrus.Infof("Portforward for %q port %v", r.GetPodSandboxId(), r.GetPort())
	defer func() {
		if err != nil {
			logrus.Errorf("Portforward for %q failed, error: %v", r.GetPodSandboxId(), err)
		} else {
			logrus.Infof("Portforward for %q returns URL %q", r.GetPodSandboxId(), res.GetUrl())
		}
	}()
	return c.CriManager.PortForward(ctx, r)
}

// UpdateRuntimeConfig updates the runtime config. Currently only handles podCIDR updates.
func (c *CriWrapper) UpdateRuntimeConfig(ctx context.Context, r *runtime.UpdateRuntimeConfigRequest) (res *runtime.UpdateRuntimeConfigResponse, err error) {
	logrus.Infof("UpdateRuntimeConfig with config %+v", r.GetRuntimeConfig())
	defer func() {
		if err != nil {
			logrus.Errorf("UpdateRuntimeConfig failed, error: %v", err)
		} else {
			logrus.Infof("UpdateRuntimeConfig returns returns successfully")
		}
	}()
	return c.CriManager.UpdateRuntimeConfig(ctx, r)
}

// Status returns the status of the runtime.
func (c *CriWrapper) Status(ctx context.Context, r *runtime.StatusRequest) (res *runtime.StatusResponse, err error) {
	logrus.Infof("Status of cri manager")
	defer func() {
		if err != nil {
			logrus.Errorf("Status failed, error: %v", err)
		} else {
			logrus.Infof("Status returns status %+v", res.GetStatus())
		}
	}()
	return c.CriManager.Status(ctx, r)
}

// ListImages lists existing images.
func (c *CriWrapper) ListImages(ctx context.Context, r *runtime.ListImagesRequest) (res *runtime.ListImagesResponse, err error) {
	logrus.Infof("ListImages with filter %+v", r.GetFilter())
	defer func() {
		if err != nil {
			logrus.Errorf("ListImages with filter %+v failed, error: %v", r.GetFilter(), err)
		} else {
			// NOTE: maybe log detailed image items with higher log level.
			logrus.Infof("ListImages with filter %+v returns image list successfully", r.GetFilter())
		}
	}()
	return c.CriManager.ListImages(ctx, r)
}

// ImageStatus returns the status of the image, returns nil if the image isn't present.
func (c *CriWrapper) ImageStatus(ctx context.Context, r *runtime.ImageStatusRequest) (res *runtime.ImageStatusResponse, err error) {
	logrus.Infof("ImageStatus for %q", r.GetImage().GetImage())
	defer func() {
		if err != nil {
			logrus.Errorf("ImageStatus for %q failed, error: %v", r.GetImage().GetImage(), err)
		} else {
			logrus.Infof("ImageStatus for %q returns image status %+v",
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
			logrus.Errorf("PullImage %q failed, error: %v", r.GetImage().GetImage(), err)
		} else {
			logrus.Infof("PullImage %q returns image reference %q",
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
			logrus.Errorf("RemoveImage %q failed, error: %v", r.GetImage().GetImage(), err)
		} else {
			logrus.Infof("RemoveImage %q returns successfully", r.GetImage().GetImage())
		}
	}()
	return c.CriManager.RemoveImage(ctx, r)
}

// ImageFsInfo returns information of the filesystem that is used to store images.
func (c *CriWrapper) ImageFsInfo(ctx context.Context, r *runtime.ImageFsInfoRequest) (res *runtime.ImageFsInfoResponse, err error) {
	logrus.Infof("ImageFsInfo of cri manager")
	defer func() {
		if err != nil {
			logrus.Errorf("ImageFsInfo failed, error: %v", err)
		} else {
			logrus.Infof("ImageFsInfo returns filesystem info %+v", res.GetImageFilesystems())
		}
	}()
	return c.CriManager.ImageFsInfo(ctx, r)
}
