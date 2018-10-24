package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/alibaba/pouch/storage/volume/types/meta"
)

// VolumeConditionType defines volume condition type.
type VolumeConditionType string

// These are valid conditions of pod.
const (
	// VolumeScheduledrepresents status of the scheduling process for this Volume.
	VolumeScheduled VolumeConditionType = "Scheduled"
	// VolumeInitialized means that all init containers in the Volume have started successfully.
	VolumeInitialized VolumeConditionType = "Initialized"
	// VolumeStopped means that all init containers in the Volume have stopped successfully.
	VolumeStopped VolumeConditionType = "Stopped"
	// VolumeStarted means that all init containers in the Volume have started successfully.
	VolumeStarted VolumeConditionType = "Started"
	// VolumeRestarted means that all init containers in the Volume have restarted successfully
	VolumeRestarted VolumeConditionType = "Restarted"
	// VolumeUpdated means that all init containers in the Volume have updated successfully
	VolumeUpdated VolumeConditionType = "Updated"
	// VolumeDeleted means that all init containers in the Volume have deleted successfully
	VolumeDeleted VolumeConditionType = "Deleted"
)

// VolumeCondition represents volume condition struct.
type VolumeCondition struct {
	Type               VolumeConditionType `json:"type"`
	Status             ConditionStatus     `json:"status"`
	LastProbeTime      *time.Time          `json:"lastProbeTime,omitempty"`
	LastTransitionTime *time.Time          `json:"lastTransitionTime,omitempty"`
	Reason             string              `json:"reason,omitempty"`
	Message            string              `json:"message,omitempty"`
	Retry              int                 `json:"retry,omitempty"`
}

// ConditionStatus string enum.
type ConditionStatus string

const (
	// ConditionTrue NodeConditionType is true.
	ConditionTrue ConditionStatus = "True"
	// ConditionFalse NodeConditionType is false.
	ConditionFalse ConditionStatus = "False"
	// ConditionUnknown NodeConditionType is Unknown.
	ConditionUnknown ConditionStatus = "Unknown"
)

// VolumePhase defines volume phase's status.
type VolumePhase string

var (
	// VolumePhasePending represents volume pending status.
	VolumePhasePending VolumePhase = "Pending"

	// VolumePhaseReady represents volume ready status.
	VolumePhaseReady VolumePhase = "Ready"

	// VolumePhaseUnknown represents volume unknown status.
	VolumePhaseUnknown VolumePhase = "Unknown"

	// VolumePhaseFailed represents volume failed status.
	VolumePhaseFailed VolumePhase = "Failed"
)

// VolumeConfig represents volume config.
type VolumeConfig struct {
	Size       string `json:"size"`
	FileSystem string `json:"filesystem"`
	MountOpt   string `json:"mountopt"`
	WriteBPS   int64  `json:"wbps"`
	ReadBPS    int64  `json:"rbps"`
	WriteIOPS  int64  `json:"wiops"`
	ReadIOPS   int64  `json:"riops"`
	TotalIOPS  int64  `json:"iops"`
}

// VolumeSpec represents volume spec.
type VolumeSpec struct {
	ClusterID     string `json:"clusterid"`
	Operable      bool   `json:"operable"`
	Backend       string `json:"backend,omitempty"`
	MountMode     string `json:"mountMode,omitempty"`
	*VolumeConfig `json:"config,inline"`
	Extra         map[string]string `json:"extra"`
}

// VolumeStatus represent volume status.
type VolumeStatus struct {
	Phase               VolumePhase       `json:"phase"`
	StartTimestamp      *time.Time        `json:"startTimestamp"`
	LastUpdateTimestamp *time.Time        `json:"lastUpdateTime"`
	Conditions          []VolumeCondition `json:"Conditions,omitempty"`
	HostIP              string            `json:"hostip,omitempty"`
	MountPoint          string            `json:"mountpath,omitempty"`
	Reason              string            `json:"reason"`
	Message             string            `json:"message"`
}

// Volume defined volume struct.
type Volume struct {
	meta.ObjectMeta `json:",inline"`
	Spec            *VolumeSpec   `json:"Spec,omitempty"`
	Status          *VolumeStatus `json:"Status,omitempty"`
}

// SetPath save the volume's path on host into volume meta data.
func (v *Volume) SetPath(p string) {
	v.Status.MountPoint = p
}

// Path return the volume's path on host.
func (v *Volume) Path() string {
	return v.Status.MountPoint
}

// Option use to get the common options or module's options by name.
func (v *Volume) Option(name string) string {
	return v.Spec.Extra[name]
}

// SetOption use to set the common options or module's options by name.
func (v *Volume) SetOption(name, value string) {
	v.Spec.Extra[name] = value
}

// Options returns all the options of volume.
func (v *Volume) Options() map[string]string {
	return v.Spec.Extra
}

// Driver return driver's name of the volume.
func (v *Volume) Driver() string {
	return v.Spec.Backend
}

// VolumeContext return volume's context.
func (v *Volume) VolumeContext() VolumeContext {
	return NewVolumeContext(v.Name, v.Driver(), v.Spec.Extra, v.Labels)
}

// Label returns volume's label.
func (v *Volume) Label(label string) string {
	return v.Labels[label]
}

// SetLabel use to set label to volume.
func (v *Volume) SetLabel(label, value string) {
	v.Labels[label] = value
}

// Size returns volume's size(bytes).
func (v *Volume) Size() string {
	if v.Spec.Size != "" {
		return v.Spec.Size
	}

	for _, k := range []string{"size", "Size", "opt.size", "opt.Size"} {
		if s, ok := v.Spec.Extra[k]; ok {
			return s
		}
	}

	return ""
}

// FileSystem returns volume's file system.
func (v *Volume) FileSystem() []string {
	return strings.Split(v.Spec.FileSystem, " ")
}

// MountOption returns volume's mount options.
func (v *Volume) MountOption() []string {
	return strings.Split(v.Spec.MountOpt, " ")
}

// Key returns the volume's name
func (v *Volume) Key() string {
	return v.Name
}

//CreateTime returns the volume's create time.
func (v *Volume) CreateTime() string {
	if v.CreationTimestamp == nil {
		return ""
	}

	return v.CreationTimestamp.Format("2006-1-2 15:04:05")
}

// VolumeContext use to define the volume's identity.
type VolumeContext struct {
	Name    string
	Driver  string
	Options map[string]string
	Labels  map[string]string
}

// NewVolumeContext returns VolumeContext instance.
func NewVolumeContext(name, driver string, options, labels map[string]string) VolumeContext {
	if options == nil {
		options = map[string]string{}
	}
	if labels == nil {
		labels = map[string]string{}
	}
	return VolumeContext{
		Name:    name,
		Driver:  driver,
		Options: options,
		Labels:  labels,
	}
}

// Equal check VolumeContext is equal or not.
func (v VolumeContext) Equal(v1 VolumeContext) bool {
	return (v.Name == v1.Name) && (v.Driver == v1.Driver)
}

// String return VolumeContext with string.
func (v VolumeContext) String() string {
	return fmt.Sprintf("<%s, %s>", v.Name, v.Driver)
}

// Invalid is used to check VolumeContext's name is valid or not.
func (v VolumeContext) Invalid() bool {
	return v.Name == ""
}
