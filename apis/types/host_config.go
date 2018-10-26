// Code generated by go-swagger; DO NOT EDIT.

package types

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"
	"strconv"

	"github.com/go-openapi/errors"
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// HostConfig Container configuration that depends on the host we are running on
// swagger:model HostConfig
type HostConfig struct {

	// Automatically remove the container when the container's process exits. This has no effect if `RestartPolicy` is set.
	AutoRemove bool `json:"AutoRemove,omitempty"`

	// A list of volume bindings for this container. Each volume binding is a string in one of these forms:
	//
	// - `host-src:container-dest` to bind-mount a host path into the container. Both `host-src`, and `container-dest` must be an _absolute_ path.
	// - `host-src:container-dest:ro` to make the bind mount read-only inside the container. Both `host-src`, and `container-dest` must be an _absolute_ path.
	// - `volume-name:container-dest` to bind-mount a volume managed by a volume driver into the container. `container-dest` must be an _absolute_ path.
	// - `volume-name:container-dest:ro` to mount the volume read-only inside the container.  `container-dest` must be an _absolute_ path.
	//
	Binds []string `json:"Binds"`

	// A list of kernel capabilities to add to the container.
	CapAdd []string `json:"CapAdd"`

	// A list of kernel capabilities to drop from the container.
	CapDrop []string `json:"CapDrop"`

	// Cgroup to use for the container.
	Cgroup string `json:"Cgroup,omitempty"`

	// Initial console size, as an `[height, width]` array. (Windows only)
	// Max Items: 2
	// Min Items: 2
	ConsoleSize []*int64 `json:"ConsoleSize"`

	// Path to a file where the container ID is written
	ContainerIDFile string `json:"ContainerIDFile,omitempty"`

	// A list of DNS servers for the container to use.
	DNS []string `json:"Dns"`

	// A list of DNS options.
	DNSOptions []string `json:"DnsOptions"`

	// A list of DNS search domains.
	DNSSearch []string `json:"DnsSearch"`

	// Whether to enable lxcfs.
	EnableLxcfs bool `json:"EnableLxcfs,omitempty"`

	// A list of hostnames/IP mappings to add to the container's `/etc/hosts` file. Specified in the form `["hostname:IP"]`.
	//
	ExtraHosts []string `json:"ExtraHosts"`

	// A list of additional groups that the container process will run as.
	GroupAdd []string `json:"GroupAdd"`

	// Initial script executed in container. The script will be executed before entrypoint or command
	InitScript string `json:"InitScript,omitempty"`

	// IPC sharing mode for the container. Possible values are:
	// - `"none"`: own private IPC namespace, with /dev/shm not mounted
	// - `"private"`: own private IPC namespace
	// - `"shareable"`: own private IPC namespace, with a possibility to share it with other containers
	// - `"container:<name|id>"`: join another (shareable) container's IPC namespace
	// - `"host"`: use the host system's IPC namespace
	// If not specified, daemon default is used, which can either be `"private"`
	// or `"shareable"`, depending on daemon version and configuration.
	//
	IpcMode string `json:"IpcMode,omitempty"`

	// Isolation technology of the container. (Windows only)
	// Enum: [default process hyperv]
	Isolation string `json:"Isolation,omitempty"`

	// A list of links for the container in the form `container_name:alias`.
	Links []string `json:"Links"`

	// The logging configuration for this container
	LogConfig *LogConfig `json:"LogConfig,omitempty"`

	// Network mode to use for this container. Supported standard values are: `bridge`, `host`, `none`, and `container:<name|id>`. Any other value is taken as a custom network's name to which this container should connect to.
	NetworkMode string `json:"NetworkMode,omitempty"`

	// An integer value containing the score given to the container in order to tune OOM killer preferences.
	// The range is in [-1000, 1000].
	//
	// Maximum: 1000
	// Minimum: -1000
	OomScoreAdj int64 `json:"OomScoreAdj,omitempty"`

	// Set the PID (Process) Namespace mode for the container. It can be either:
	// - `"container:<name|id>"`: joins another container's PID namespace
	// - `"host"`: use the host's PID namespace inside the container
	//
	PidMode string `json:"PidMode,omitempty"`

	// A map of exposed container ports and the host port they should map to.
	PortBindings PortMap `json:"PortBindings,omitempty"`

	// Gives the container full access to the host.
	Privileged bool `json:"Privileged,omitempty"`

	// Allocates a random host port for all of a container's exposed ports.
	PublishAllPorts bool `json:"PublishAllPorts,omitempty"`

	// Mount the container's root filesystem as read only.
	ReadonlyRootfs bool `json:"ReadonlyRootfs,omitempty"`

	// Restart policy to be used to manage the container
	RestartPolicy *RestartPolicy `json:"RestartPolicy,omitempty"`

	// Whether to start container in rich container mode. (default false)
	Rich bool `json:"Rich,omitempty"`

	// Choose one rich container mode.(default dumb-init)
	// Enum: [dumb-init sbin-init systemd]
	RichMode string `json:"RichMode,omitempty"`

	// Runtime to use with this container.
	Runtime string `json:"Runtime,omitempty"`

	// A list of string values to customize labels for MLS systems, such as SELinux.
	SecurityOpt []string `json:"SecurityOpt"`

	// Shim name, pass to containerd to choose which shim is used for container.
	//
	Shim string `json:"Shim,omitempty"`

	// Size of `/dev/shm` in bytes. If omitted, the system uses 64MB.
	// Minimum: 0
	ShmSize *int64 `json:"ShmSize,omitempty"`

	// Storage driver options for this container, in the form `{"size": "120G"}`.
	//
	StorageOpt map[string]string `json:"StorageOpt,omitempty"`

	// A list of kernel parameters (sysctls) to set in the container. For example: `{"net.ipv4.ip_forward": "1"}`
	//
	Sysctls map[string]string `json:"Sysctls,omitempty"`

	// A map of container directories which should be replaced by tmpfs mounts, and their corresponding mount options. For example: `{ "/run": "rw,noexec,nosuid,size=65536k" }`.
	//
	Tmpfs map[string]string `json:"Tmpfs,omitempty"`

	// UTS namespace to use for the container.
	UTSMode string `json:"UTSMode,omitempty"`

	// Sets the usernamespace mode for the container when usernamespace remapping option is enabled.
	UsernsMode string `json:"UsernsMode,omitempty"`

	// Driver that this container uses to mount volumes.
	VolumeDriver string `json:"VolumeDriver,omitempty"`

	// A list of volumes to inherit from another container, specified in the form `<container name>[:<ro|rw>]`.
	VolumesFrom []string `json:"VolumesFrom"`

	Resources
}

// UnmarshalJSON unmarshals this object from a JSON structure
func (m *HostConfig) UnmarshalJSON(raw []byte) error {
	// AO0
	var dataAO0 struct {
		AutoRemove bool `json:"AutoRemove,omitempty"`

		Binds []string `json:"Binds"`

		CapAdd []string `json:"CapAdd"`

		CapDrop []string `json:"CapDrop"`

		Cgroup string `json:"Cgroup,omitempty"`

		ConsoleSize []*int64 `json:"ConsoleSize"`

		ContainerIDFile string `json:"ContainerIDFile,omitempty"`

		DNS []string `json:"Dns"`

		DNSOptions []string `json:"DnsOptions"`

		DNSSearch []string `json:"DnsSearch"`

		EnableLxcfs bool `json:"EnableLxcfs,omitempty"`

		ExtraHosts []string `json:"ExtraHosts"`

		GroupAdd []string `json:"GroupAdd"`

		InitScript string `json:"InitScript,omitempty"`

		IpcMode string `json:"IpcMode,omitempty"`

		Isolation string `json:"Isolation,omitempty"`

		Links []string `json:"Links"`

		LogConfig *LogConfig `json:"LogConfig,omitempty"`

		NetworkMode string `json:"NetworkMode,omitempty"`

		OomScoreAdj int64 `json:"OomScoreAdj,omitempty"`

		PidMode string `json:"PidMode,omitempty"`

		PortBindings PortMap `json:"PortBindings,omitempty"`

		Privileged bool `json:"Privileged,omitempty"`

		PublishAllPorts bool `json:"PublishAllPorts,omitempty"`

		ReadonlyRootfs bool `json:"ReadonlyRootfs,omitempty"`

		RestartPolicy *RestartPolicy `json:"RestartPolicy,omitempty"`

		Rich bool `json:"Rich,omitempty"`

		RichMode string `json:"RichMode,omitempty"`

		Runtime string `json:"Runtime,omitempty"`

		SecurityOpt []string `json:"SecurityOpt"`

		Shim string `json:"Shim,omitempty"`

		ShmSize *int64 `json:"ShmSize,omitempty"`

		StorageOpt map[string]string `json:"StorageOpt,omitempty"`

		Sysctls map[string]string `json:"Sysctls,omitempty"`

		Tmpfs map[string]string `json:"Tmpfs,omitempty"`

		UTSMode string `json:"UTSMode,omitempty"`

		UsernsMode string `json:"UsernsMode,omitempty"`

		VolumeDriver string `json:"VolumeDriver,omitempty"`

		VolumesFrom []string `json:"VolumesFrom"`
	}
	if err := swag.ReadJSON(raw, &dataAO0); err != nil {
		return err
	}

	m.AutoRemove = dataAO0.AutoRemove

	m.Binds = dataAO0.Binds

	m.CapAdd = dataAO0.CapAdd

	m.CapDrop = dataAO0.CapDrop

	m.Cgroup = dataAO0.Cgroup

	m.ConsoleSize = dataAO0.ConsoleSize

	m.ContainerIDFile = dataAO0.ContainerIDFile

	m.DNS = dataAO0.DNS

	m.DNSOptions = dataAO0.DNSOptions

	m.DNSSearch = dataAO0.DNSSearch

	m.EnableLxcfs = dataAO0.EnableLxcfs

	m.ExtraHosts = dataAO0.ExtraHosts

	m.GroupAdd = dataAO0.GroupAdd

	m.InitScript = dataAO0.InitScript

	m.IpcMode = dataAO0.IpcMode

	m.Isolation = dataAO0.Isolation

	m.Links = dataAO0.Links

	m.LogConfig = dataAO0.LogConfig

	m.NetworkMode = dataAO0.NetworkMode

	m.OomScoreAdj = dataAO0.OomScoreAdj

	m.PidMode = dataAO0.PidMode

	m.PortBindings = dataAO0.PortBindings

	m.Privileged = dataAO0.Privileged

	m.PublishAllPorts = dataAO0.PublishAllPorts

	m.ReadonlyRootfs = dataAO0.ReadonlyRootfs

	m.RestartPolicy = dataAO0.RestartPolicy

	m.Rich = dataAO0.Rich

	m.RichMode = dataAO0.RichMode

	m.Runtime = dataAO0.Runtime

	m.SecurityOpt = dataAO0.SecurityOpt

	m.Shim = dataAO0.Shim

	m.ShmSize = dataAO0.ShmSize

	m.StorageOpt = dataAO0.StorageOpt

	m.Sysctls = dataAO0.Sysctls

	m.Tmpfs = dataAO0.Tmpfs

	m.UTSMode = dataAO0.UTSMode

	m.UsernsMode = dataAO0.UsernsMode

	m.VolumeDriver = dataAO0.VolumeDriver

	m.VolumesFrom = dataAO0.VolumesFrom

	// AO1
	var aO1 Resources
	if err := swag.ReadJSON(raw, &aO1); err != nil {
		return err
	}
	m.Resources = aO1

	return nil
}

// MarshalJSON marshals this object to a JSON structure
func (m HostConfig) MarshalJSON() ([]byte, error) {
	_parts := make([][]byte, 0, 2)

	var dataAO0 struct {
		AutoRemove bool `json:"AutoRemove,omitempty"`

		Binds []string `json:"Binds"`

		CapAdd []string `json:"CapAdd"`

		CapDrop []string `json:"CapDrop"`

		Cgroup string `json:"Cgroup,omitempty"`

		ConsoleSize []*int64 `json:"ConsoleSize"`

		ContainerIDFile string `json:"ContainerIDFile,omitempty"`

		DNS []string `json:"Dns"`

		DNSOptions []string `json:"DnsOptions"`

		DNSSearch []string `json:"DnsSearch"`

		EnableLxcfs bool `json:"EnableLxcfs,omitempty"`

		ExtraHosts []string `json:"ExtraHosts"`

		GroupAdd []string `json:"GroupAdd"`

		InitScript string `json:"InitScript,omitempty"`

		IpcMode string `json:"IpcMode,omitempty"`

		Isolation string `json:"Isolation,omitempty"`

		Links []string `json:"Links"`

		LogConfig *LogConfig `json:"LogConfig,omitempty"`

		NetworkMode string `json:"NetworkMode,omitempty"`

		OomScoreAdj int64 `json:"OomScoreAdj,omitempty"`

		PidMode string `json:"PidMode,omitempty"`

		PortBindings PortMap `json:"PortBindings,omitempty"`

		Privileged bool `json:"Privileged,omitempty"`

		PublishAllPorts bool `json:"PublishAllPorts,omitempty"`

		ReadonlyRootfs bool `json:"ReadonlyRootfs,omitempty"`

		RestartPolicy *RestartPolicy `json:"RestartPolicy,omitempty"`

		Rich bool `json:"Rich,omitempty"`

		RichMode string `json:"RichMode,omitempty"`

		Runtime string `json:"Runtime,omitempty"`

		SecurityOpt []string `json:"SecurityOpt"`

		Shim string `json:"Shim,omitempty"`

		ShmSize *int64 `json:"ShmSize,omitempty"`

		StorageOpt map[string]string `json:"StorageOpt,omitempty"`

		Sysctls map[string]string `json:"Sysctls,omitempty"`

		Tmpfs map[string]string `json:"Tmpfs,omitempty"`

		UTSMode string `json:"UTSMode,omitempty"`

		UsernsMode string `json:"UsernsMode,omitempty"`

		VolumeDriver string `json:"VolumeDriver,omitempty"`

		VolumesFrom []string `json:"VolumesFrom"`
	}

	dataAO0.AutoRemove = m.AutoRemove

	dataAO0.Binds = m.Binds

	dataAO0.CapAdd = m.CapAdd

	dataAO0.CapDrop = m.CapDrop

	dataAO0.Cgroup = m.Cgroup

	dataAO0.ConsoleSize = m.ConsoleSize

	dataAO0.ContainerIDFile = m.ContainerIDFile

	dataAO0.DNS = m.DNS

	dataAO0.DNSOptions = m.DNSOptions

	dataAO0.DNSSearch = m.DNSSearch

	dataAO0.EnableLxcfs = m.EnableLxcfs

	dataAO0.ExtraHosts = m.ExtraHosts

	dataAO0.GroupAdd = m.GroupAdd

	dataAO0.InitScript = m.InitScript

	dataAO0.IpcMode = m.IpcMode

	dataAO0.Isolation = m.Isolation

	dataAO0.Links = m.Links

	dataAO0.LogConfig = m.LogConfig

	dataAO0.NetworkMode = m.NetworkMode

	dataAO0.OomScoreAdj = m.OomScoreAdj

	dataAO0.PidMode = m.PidMode

	dataAO0.PortBindings = m.PortBindings

	dataAO0.Privileged = m.Privileged

	dataAO0.PublishAllPorts = m.PublishAllPorts

	dataAO0.ReadonlyRootfs = m.ReadonlyRootfs

	dataAO0.RestartPolicy = m.RestartPolicy

	dataAO0.Rich = m.Rich

	dataAO0.RichMode = m.RichMode

	dataAO0.Runtime = m.Runtime

	dataAO0.SecurityOpt = m.SecurityOpt

	dataAO0.Shim = m.Shim

	dataAO0.ShmSize = m.ShmSize

	dataAO0.StorageOpt = m.StorageOpt

	dataAO0.Sysctls = m.Sysctls

	dataAO0.Tmpfs = m.Tmpfs

	dataAO0.UTSMode = m.UTSMode

	dataAO0.UsernsMode = m.UsernsMode

	dataAO0.VolumeDriver = m.VolumeDriver

	dataAO0.VolumesFrom = m.VolumesFrom

	jsonDataAO0, errAO0 := swag.WriteJSON(dataAO0)
	if errAO0 != nil {
		return nil, errAO0
	}
	_parts = append(_parts, jsonDataAO0)

	aO1, err := swag.WriteJSON(m.Resources)
	if err != nil {
		return nil, err
	}
	_parts = append(_parts, aO1)

	return swag.ConcatJSON(_parts...), nil
}

// Validate validates this host config
func (m *HostConfig) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateConsoleSize(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateIsolation(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateLogConfig(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateOomScoreAdj(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validatePortBindings(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRestartPolicy(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateRichMode(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateShmSize(formats); err != nil {
		res = append(res, err)
	}

	// validation for a type composition with Resources
	if err := m.Resources.Validate(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *HostConfig) validateConsoleSize(formats strfmt.Registry) error {

	if swag.IsZero(m.ConsoleSize) { // not required
		return nil
	}

	iConsoleSizeSize := int64(len(m.ConsoleSize))

	if err := validate.MinItems("ConsoleSize", "body", iConsoleSizeSize, 2); err != nil {
		return err
	}

	if err := validate.MaxItems("ConsoleSize", "body", iConsoleSizeSize, 2); err != nil {
		return err
	}

	for i := 0; i < len(m.ConsoleSize); i++ {
		if swag.IsZero(m.ConsoleSize[i]) { // not required
			continue
		}

		if err := validate.MinimumInt("ConsoleSize"+"."+strconv.Itoa(i), "body", int64(*m.ConsoleSize[i]), 0, false); err != nil {
			return err
		}

	}

	return nil
}

var hostConfigTypeIsolationPropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["default","process","hyperv"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		hostConfigTypeIsolationPropEnum = append(hostConfigTypeIsolationPropEnum, v)
	}
}

// property enum
func (m *HostConfig) validateIsolationEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, hostConfigTypeIsolationPropEnum); err != nil {
		return err
	}
	return nil
}

func (m *HostConfig) validateIsolation(formats strfmt.Registry) error {

	if swag.IsZero(m.Isolation) { // not required
		return nil
	}

	// value enum
	if err := m.validateIsolationEnum("Isolation", "body", m.Isolation); err != nil {
		return err
	}

	return nil
}

func (m *HostConfig) validateLogConfig(formats strfmt.Registry) error {

	if swag.IsZero(m.LogConfig) { // not required
		return nil
	}

	if m.LogConfig != nil {
		if err := m.LogConfig.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("LogConfig")
			}
			return err
		}
	}

	return nil
}

func (m *HostConfig) validateOomScoreAdj(formats strfmt.Registry) error {

	if swag.IsZero(m.OomScoreAdj) { // not required
		return nil
	}

	if err := validate.MinimumInt("OomScoreAdj", "body", int64(m.OomScoreAdj), -1000, false); err != nil {
		return err
	}

	if err := validate.MaximumInt("OomScoreAdj", "body", int64(m.OomScoreAdj), 1000, false); err != nil {
		return err
	}

	return nil
}

func (m *HostConfig) validatePortBindings(formats strfmt.Registry) error {

	if swag.IsZero(m.PortBindings) { // not required
		return nil
	}

	if err := m.PortBindings.Validate(formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("PortBindings")
		}
		return err
	}

	return nil
}

func (m *HostConfig) validateRestartPolicy(formats strfmt.Registry) error {

	if swag.IsZero(m.RestartPolicy) { // not required
		return nil
	}

	if m.RestartPolicy != nil {
		if err := m.RestartPolicy.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("RestartPolicy")
			}
			return err
		}
	}

	return nil
}

var hostConfigTypeRichModePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["dumb-init","sbin-init","systemd"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		hostConfigTypeRichModePropEnum = append(hostConfigTypeRichModePropEnum, v)
	}
}

// property enum
func (m *HostConfig) validateRichModeEnum(path, location string, value string) error {
	if err := validate.Enum(path, location, value, hostConfigTypeRichModePropEnum); err != nil {
		return err
	}
	return nil
}

func (m *HostConfig) validateRichMode(formats strfmt.Registry) error {

	if swag.IsZero(m.RichMode) { // not required
		return nil
	}

	// value enum
	if err := m.validateRichModeEnum("RichMode", "body", m.RichMode); err != nil {
		return err
	}

	return nil
}

func (m *HostConfig) validateShmSize(formats strfmt.Registry) error {

	if swag.IsZero(m.ShmSize) { // not required
		return nil
	}

	if err := validate.MinimumInt("ShmSize", "body", int64(*m.ShmSize), 0, false); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *HostConfig) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *HostConfig) UnmarshalBinary(b []byte) error {
	var res HostConfig
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
