package mgr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/opts"
	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/ctrd"

	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/devices"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// FIXME: these variables have no relation with spec, move them.
const (
	// ProfileNamePrefix is the prefix for loading profiles on a localhost. Eg. localhost/profileName.
	ProfileNamePrefix = "localhost/"
	// ProfileRuntimeDefault indicates that we should use or create a runtime default profile.
	ProfileRuntimeDefault = "runtime/default"
	// ProfileDockerDefault indicates that we should use or create a docker default profile.
	ProfileDockerDefault = "docker/default"
	// ProfilePouchDefault indicates that we should use or create a pouch default profile.
	ProfilePouchDefault = "pouch/default"
	// ProfileNameUnconfined is a string indicating one should run a pod/containerd without a security profile.
	ProfileNameUnconfined = "unconfined"

	// defaultCgroupParent is default cgroup parent.
	defaultCgroupParent = "pouch"
)

// Setup linux-platform-sepecific specification.
func populatePlatform(ctx context.Context, c *Container, specWrapper *SpecWrapper) error {
	s := specWrapper.s
	if s.Linux == nil {
		s.Linux = &specs.Linux{}
	}

	// same with containerd use. or make it a variable
	// set default cgroup parent
	cgroupsParent := "/default"
	if specWrapper.useSystemd {
		cgroupsParent = "system.slice"
	}

	if c.HostConfig.CgroupParent != "" {
		cgroupsParent = filepath.Clean(c.HostConfig.CgroupParent)
	}

	if specWrapper.useSystemd {
		s.Linux.CgroupsPath = cgroupsParent + ":" + defaultCgroupParent + ":" + c.ID
	} else {
		s.Linux.CgroupsPath = filepath.Join(cgroupsParent, c.ID)
	}

	s.Linux.Sysctl = c.HostConfig.Sysctls

	if c.HostConfig.IntelRdtL3Cbm != "" {
		s.Linux.IntelRdt = &specs.LinuxIntelRdt{
			L3CacheSchema: c.HostConfig.IntelRdtL3Cbm,
		}
	}

	// setup something depend on privileged authority
	if !c.HostConfig.Privileged {
		s.Linux.MountLabel = c.MountLabel
	} else {
		s.Linux.ReadonlyPaths = nil
		s.Linux.MaskedPaths = nil
	}

	// Apply masked paths if specified.
	if c.MaskedPaths != nil {
		s.Linux.MaskedPaths = make([]string, len(c.MaskedPaths))
		copy(s.Linux.MaskedPaths, c.MaskedPaths)
	}

	// Apply readonly paths if specified.
	if c.ReadonlyPaths != nil {
		s.Linux.ReadonlyPaths = make([]string, len(c.ReadonlyPaths))
		copy(s.Linux.ReadonlyPaths, c.ReadonlyPaths)
	}

	// start to setup linux seccomp
	if err := setupSeccomp(ctx, c, s); err != nil {
		return err
	}

	// start to setup linux resource
	if err := setupResource(ctx, c, s); err != nil {
		return err
	}

	return setupNamespaces(ctx, c, specWrapper)
}

// setupResource creates linux resource spec.
func setupResource(ctx context.Context, c *Container, s *specs.Spec) error {
	if s.Linux.Resources == nil {
		s.Linux.Resources = &specs.LinuxResources{}
	}

	// start to setup cpu and memory cgroup
	setupCPU(ctx, c.HostConfig.Resources, s)
	setupMemory(ctx, c.HostConfig.Resources, s)

	// start to setup blkio cgroup
	if err := setupBlkio(ctx, c.HostConfig.Resources, s); err != nil {
		return err
	}

	// start to setup device cgroup
	if err := setupDevices(ctx, c, s); err != nil {
		return err
	}

	// start to setup pids limit
	s.Linux.Resources.Pids = &specs.LinuxPids{
		Limit: c.HostConfig.PidsLimit,
	}

	return nil
}

// setupResource creates linux blkio resource spec.
func setupBlkio(ctx context.Context, r types.Resources, s *specs.Spec) error {
	weightDevice, err := ctrd.GetWeightDevice(r.BlkioWeightDevice)
	if err != nil {
		return err
	}
	readBpsDevice, err := ctrd.GetThrottleDevice(r.BlkioDeviceReadBps)
	if err != nil {
		return err
	}
	writeBpsDevice, err := ctrd.GetThrottleDevice(r.BlkioDeviceWriteBps)
	if err != nil {
		return err
	}
	readIOpsDevice, err := ctrd.GetThrottleDevice(r.BlkioDeviceReadIOps)
	if err != nil {
		return err
	}
	writeIOpsDevice, err := ctrd.GetThrottleDevice(r.BlkioDeviceWriteIOps)
	if err != nil {
		return err
	}

	s.Linux.Resources.BlockIO = &specs.LinuxBlockIO{
		Weight:                  &r.BlkioWeight,
		WeightDevice:            weightDevice,
		ThrottleReadBpsDevice:   readBpsDevice,
		ThrottleReadIOPSDevice:  readIOpsDevice,
		ThrottleWriteBpsDevice:  writeBpsDevice,
		ThrottleWriteIOPSDevice: writeIOpsDevice,
	}

	return nil
}

// setupResource creates linux cpu resource spec
func setupCPU(ctx context.Context, r types.Resources, s *specs.Spec) {
	cpu := &specs.LinuxCPU{
		Cpus: r.CpusetCpus,
		Mems: r.CpusetMems,
	}

	if r.CPUShares != 0 {
		v := uint64(r.CPUShares)
		cpu.Shares = &v
	}

	if r.CPUPeriod != 0 {
		v := uint64(r.CPUPeriod)
		cpu.Period = &v
	}

	if r.CPUQuota != 0 {
		v := int64(r.CPUQuota)
		cpu.Quota = &v
	}

	s.Linux.Resources.CPU = cpu
}

// setupResource creates linux memory resource spec.
func setupMemory(ctx context.Context, r types.Resources, s *specs.Spec) {
	memory := &specs.LinuxMemory{}
	if r.Memory > 0 {
		v := r.Memory
		memory.Limit = &v
	}

	if r.MemorySwap != 0 {
		v := r.MemorySwap
		memory.Swap = &v
	}

	if r.MemorySwappiness != nil && *r.MemorySwappiness != -1 {
		v := uint64(*r.MemorySwappiness)
		memory.Swappiness = &v
	}

	if r.OomKillDisable != nil {
		v := bool(*r.OomKillDisable)
		memory.DisableOOMKiller = &v
	}

	s.Linux.Resources.Memory = memory
}

// setupResource creates linux device resource spec.
func setupDevices(ctx context.Context, c *Container, s *specs.Spec) error {
	var devs []specs.LinuxDevice
	devPermissions := s.Linux.Resources.Devices
	if c.HostConfig.Privileged {
		hostDevices, err := devices.HostDevices()
		if err != nil {
			return err
		}
		for _, d := range hostDevices {
			devs = append(devs, linuxDevice(d))
		}
		devPermissions = []specs.LinuxDeviceCgroup{
			{
				Allow:  true,
				Access: "rwm",
			},
		}
	} else {
		for _, deviceMapping := range c.HostConfig.Devices {
			if !opts.ValidateDeviceMode(deviceMapping.CgroupPermissions) {
				return fmt.Errorf("%s invalid device mode: %s", deviceMapping.PathOnHost, deviceMapping.CgroupPermissions)
			}
			d, dPermissions, err := devicesFromPath(deviceMapping.PathOnHost, deviceMapping.PathInContainer, deviceMapping.CgroupPermissions)
			if err != nil {
				return err
			}
			devs = append(devs, d...)
			devPermissions = append(devPermissions, dPermissions...)
		}
	}

	s.Linux.Devices = append(s.Linux.Devices, devs...)
	s.Linux.Resources.Devices = devPermissions
	return nil
}

func u32Ptr(i int64) *uint32     { u := uint32(i); return &u }
func fmPtr(i int64) *os.FileMode { fm := os.FileMode(i); return &fm }

// linuxDevice convert a libcontainer configs.Device to a specs.LinuxDevice object.
func linuxDevice(d *configs.Device) specs.LinuxDevice {
	return specs.LinuxDevice{
		Type:     string(d.Type),
		Path:     d.Path,
		Major:    d.Major,
		Minor:    d.Minor,
		FileMode: fmPtr(int64(d.FileMode)),
		UID:      u32Ptr(int64(d.Uid)),
		GID:      u32Ptr(int64(d.Gid)),
	}
}

func deviceCgroup(d *configs.Device) specs.LinuxDeviceCgroup {
	t := string(d.Type)
	return specs.LinuxDeviceCgroup{
		Allow:  true,
		Type:   t,
		Major:  &d.Major,
		Minor:  &d.Minor,
		Access: d.Permissions,
	}
}

func devicesFromPath(pathOnHost, pathInContainer, cgroupPermissions string) (devs []specs.LinuxDevice, devPermissions []specs.LinuxDeviceCgroup, err error) {
	resolvedPathOnHost := pathOnHost

	// check if it is a symbolic link
	if src, e := os.Lstat(pathOnHost); e == nil && src.Mode()&os.ModeSymlink == os.ModeSymlink {
		if linkedPathOnHost, e := filepath.EvalSymlinks(pathOnHost); e == nil {
			resolvedPathOnHost = linkedPathOnHost
		}
	}

	device, err := devices.DeviceFromPath(resolvedPathOnHost, cgroupPermissions)
	if err == nil {
		device.Path = pathInContainer
		return append(devs, linuxDevice(device)), append(devPermissions, deviceCgroup(device)), nil
	}

	// if the device is not a device node
	// try to see if it's a directory holding many devices
	if err == devices.ErrNotADevice {

		// check if it is a directory
		if src, e := os.Stat(resolvedPathOnHost); e == nil && src.IsDir() {

			// mount the internal devices recursively
			filepath.Walk(resolvedPathOnHost, func(dpath string, f os.FileInfo, e error) error {
				childDevice, e := devices.DeviceFromPath(dpath, cgroupPermissions)
				if e != nil {
					// ignore the device
					return nil
				}

				// add the device to userSpecified devices
				childDevice.Path = strings.Replace(dpath, resolvedPathOnHost, pathInContainer, 1)
				devs = append(devs, linuxDevice(childDevice))
				devPermissions = append(devPermissions, deviceCgroup(childDevice))

				return nil
			})
		}
	}

	if len(devs) > 0 {
		return devs, devPermissions, nil
	}

	return devs, devPermissions, fmt.Errorf("error gathering device information while adding custom device %q: %s", pathOnHost, err)
}

// setupNamespaces creates linux namespaces spec.
func setupNamespaces(ctx context.Context, c *Container, specWrapper *SpecWrapper) error {
	// create user namespace spec
	if err := setupUserNamespace(ctx, c, specWrapper); err != nil {
		return err
	}

	// create network namespace spec
	if err := setupNetworkNamespace(ctx, c, specWrapper); err != nil {
		return err
	}

	// create ipc namespace spec
	if err := setupIpcNamespace(ctx, c, specWrapper); err != nil {
		return err
	}

	// create pid namespace spec
	if err := setupPidNamespace(ctx, c, specWrapper); err != nil {
		return err
	}

	// create uts namespace spec
	return setupUtsNamespace(ctx, c, specWrapper)
}

// isHost indicates whether the container shares the host's corresponding namespace.
func isHost(mode string) bool {
	return mode == "host"
}

// isContainer indicates whether the container uses another container's corresponding namespace.
func isContainer(mode string) bool {
	parts := strings.SplitN(mode, ":", 2)
	return len(parts) > 1 && parts[0] == "container"
}

// connectedContainer is the id or name of the container whose namespace this container share with.
func connectedContainer(mode string) string {
	parts := strings.SplitN(mode, ":", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func getIpcContainer(ctx context.Context, mgr ContainerMgr, id string) (*Container, error) {
	// Check whether the container exists.
	c, err := mgr.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("can't join IPC namespace of container %q: %v", id, err)
	}

	// TODO: check whether the container is running and not restarting.

	// TODO: check whether the container's ipc namespace is shareable.

	return c, nil
}

func getPidContainer(ctx context.Context, mgr ContainerMgr, id string) (*Container, error) {
	// Check the container exists.
	c, err := mgr.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("can't join PID namespace of %q: %v", id, err)
	}

	// TODO: check whether the container is running and not restarting.

	return c, nil
}

// TODO
func setupUserNamespace(ctx context.Context, c *Container, specWrapper *SpecWrapper) error {
	return nil
}

func setupNetworkNamespace(ctx context.Context, c *Container, specWrapper *SpecWrapper) error {
	if c.Config.NetworkDisabled {
		return nil
	}

	s := specWrapper.s
	ns := specs.LinuxNamespace{Type: specs.NetworkNamespace}

	networkMode := c.HostConfig.NetworkMode
	if IsContainer(networkMode) {
		origContainer, err := specWrapper.ctrMgr.Get(ctx, strings.SplitN(networkMode, ":", 2)[1])
		if err != nil {
			return err
		}
		if c.ID == origContainer.ID {
			return fmt.Errorf("can not join own network")
		} else if origContainer.State.Status != types.StatusRunning {
			return fmt.Errorf("can not join network of a non running container: %s", origContainer.ID)
		}

		ns.Path = fmt.Sprintf("/proc/%d/ns/net", origContainer.State.Pid)
	} else if IsHost(networkMode) {
		ns.Path = c.NetworkSettings.SandboxKey
	}
	setNamespace(s, ns)

	for _, ns := range s.Linux.Namespaces {
		if ns.Type == "network" && ns.Path == "" && !c.Config.NetworkDisabled {
			target, err := os.Readlink(filepath.Join("/proc", strconv.Itoa(os.Getpid()), "exe"))
			if err != nil {
				return err
			}

			netnsPrestart := specs.Hook{
				Path: target,
				Args: []string{"libnetwork-setkey", c.ID, specWrapper.netMgr.Controller().ID()},
			}
			s.Hooks.Prestart = append(s.Hooks.Prestart, netnsPrestart)
		}
	}
	return nil
}

func setupIpcNamespace(ctx context.Context, c *Container, specWrapper *SpecWrapper) error {
	s := specWrapper.s
	ipcMode := c.HostConfig.IpcMode
	switch {
	case isContainer(ipcMode):
		ns := specs.LinuxNamespace{Type: specs.IPCNamespace}
		c, err := getIpcContainer(ctx, specWrapper.ctrMgr, connectedContainer(ipcMode))
		if err != nil {
			return fmt.Errorf("setup container ipc namespace mode failed: %v", err)
		}
		ns.Path = fmt.Sprintf("/proc/%d/ns/ipc", c.State.Pid)
		setNamespace(s, ns)
	case isHost(ipcMode):
		removeNamespace(s, specs.IPCNamespace)
	default:
		ns := specs.LinuxNamespace{Type: specs.IPCNamespace}
		setNamespace(s, ns)
	}
	return nil
}

func setupPidNamespace(ctx context.Context, c *Container, specWrapper *SpecWrapper) error {
	s := specWrapper.s
	pidMode := c.HostConfig.PidMode
	switch {
	case isContainer(pidMode):
		ns := specs.LinuxNamespace{Type: specs.PIDNamespace}
		c, err := getPidContainer(ctx, specWrapper.ctrMgr, connectedContainer(pidMode))
		if err != nil {
			return fmt.Errorf("setup container pid namespace mode failed: %v", err)
		}
		ns.Path = fmt.Sprintf("/proc/%d/ns/pid", c.State.Pid)
		setNamespace(s, ns)
	case isHost(pidMode):
		removeNamespace(s, specs.PIDNamespace)
	default:
		ns := specs.LinuxNamespace{Type: specs.PIDNamespace}
		setNamespace(s, ns)
	}
	return nil
}

func setupUtsNamespace(ctx context.Context, c *Container, specWrapper *SpecWrapper) error {
	s := specWrapper.s
	utsMode := c.HostConfig.UTSMode
	switch {
	case isHost(utsMode):
		removeNamespace(s, specs.UTSNamespace)
		// remove hostname
		s.Hostname = ""
	default:
		ns := specs.LinuxNamespace{Type: specs.UTSNamespace}
		setNamespace(s, ns)
	}
	return nil
}

func setNamespace(s *specs.Spec, ns specs.LinuxNamespace) {
	for i, n := range s.Linux.Namespaces {
		if n.Type == ns.Type {
			s.Linux.Namespaces[i] = ns
			return
		}
	}
	s.Linux.Namespaces = append(s.Linux.Namespaces, ns)
}

func removeNamespace(s *specs.Spec, nsType specs.LinuxNamespaceType) {
	for i, n := range s.Linux.Namespaces {
		if n.Type == nsType {
			s.Linux.Namespaces = append(s.Linux.Namespaces[:i], s.Linux.Namespaces[i+1:]...)
			return
		}
	}
}
