package metrics

import (
	"sync"

	util_metrics "github.com/alibaba/pouch/pkg/utils/metrics"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
)

func init() {
	// Register prometheus metrics.
	Register()
}

const (
	subsystemCRI = "cri"
)

var (
	// GRPCMetrics create some standard server metrics.
	GRPCMetrics = grpc_prometheus.NewServerMetrics()

	// PodActionsCounter records the number of pod operations.
	PodActionsCounter = util_metrics.NewLabelCounter(subsystemCRI, "pod_actions_counter", "The number of pod operations", "action")

	// PodSuccessActionsCounter records the number of pod success operations.
	PodSuccessActionsCounter = util_metrics.NewLabelCounter(subsystemCRI, "pod_success_actions_counter", "The number of pod success operations", "action")

	// PodActionsTimer records the time cost of each pod action.
	PodActionsTimer = util_metrics.NewLabelTimer(subsystemCRI, "pod_actions", "The number of seconds it takes to process each pod action", "action")

	// ContainerActionsCounter records the number of container operations.
	ContainerActionsCounter = util_metrics.NewLabelCounter(subsystemCRI, "container_actions_counter", "The number of container operations", "action")

	// ContainerSuccessActionsCounter records the number of container success operations.
	ContainerSuccessActionsCounter = util_metrics.NewLabelCounter(subsystemCRI, "container_success_actions_counter", "The number of container success operations", "action")

	// ContainerActionsTimer records the time cost of each container action.
	ContainerActionsTimer = util_metrics.NewLabelTimer(subsystemCRI, "container_actions", "The number of seconds it takes to process each container action", "action")

	// ImagePullSummary records the summary of pulling image latency.
	ImagePullSummary = util_metrics.NewLabelSummary(subsystemCRI, "image_pull_latency_microseconds", "Latency in microseconds to pull a image.", "image")

	// ImageActionsCounter records the number of image operations.
	ImageActionsCounter = util_metrics.NewLabelCounter(subsystemCRI, "image_actions_counter", "The number of image operations", "action")

	// ImageSuccessActionsCounter the number of image success operations.
	ImageSuccessActionsCounter = util_metrics.NewLabelCounter(subsystemCRI, "image_success_actions_counter", "The number of image success operations", "action")

	// ImageActionsTimer records the time cost of each image action.
	ImageActionsTimer = util_metrics.NewLabelTimer(subsystemCRI, "image_actions", "The number of seconds it takes to process each image action", "action")

	// VolumeActionsCounter records the number of volume operations.
	VolumeActionsCounter = util_metrics.NewLabelCounter(subsystemCRI, "volume_actions_counter", "The number of volume operations", "action")

	// VolumeSuccessActionsCounter the number of volume success operations.
	VolumeSuccessActionsCounter = util_metrics.NewLabelCounter(subsystemCRI, "volume_success_actions_counter", "The number of volume success operations", "action")

	// VolumeActionsTimer records the time cost of each volume action.
	VolumeActionsTimer = util_metrics.NewLabelTimer(subsystemCRI, "volume_actions", "The number of seconds it takes to process each volume action", "action")

	// RuntimeActionsCounter records the number of runtime operations.
	RuntimeActionsCounter = util_metrics.NewLabelCounter(subsystemCRI, "runtime_actions_counter", "The number of runtime operations", "action")

	// RuntimeSuccessActionsCounter the number of runtime success operations.
	RuntimeSuccessActionsCounter = util_metrics.NewLabelCounter(subsystemCRI, "runtime_success_actions_counter", "The number of runtime success operations", "action")

	// RuntimeActionsTimer records the time cost of each runtime action.
	RuntimeActionsTimer = util_metrics.NewLabelTimer(subsystemCRI, "runtime_actions", "The number of seconds it takes to process each runtime action", "action")
)

var registerMetrics sync.Once

// Register all metrics.
func Register() {
	// Get a custom prometheus registry.
	registry := util_metrics.GetCustomPrometheusRegistry()
	registerMetrics.Do(func() {
		// Register the custom metrics.
		registry.MustRegister(PodActionsCounter)
		registry.MustRegister(PodSuccessActionsCounter)
		registry.MustRegister(PodActionsTimer)
		registry.MustRegister(ContainerActionsCounter)
		registry.MustRegister(ContainerSuccessActionsCounter)
		registry.MustRegister(ContainerActionsTimer)
		registry.MustRegister(ImagePullSummary)
		registry.MustRegister(ImageActionsCounter)
		registry.MustRegister(ImageSuccessActionsCounter)
		registry.MustRegister(ImageActionsTimer)
		registry.MustRegister(VolumeActionsCounter)
		registry.MustRegister(VolumeSuccessActionsCounter)
		registry.MustRegister(VolumeActionsTimer)
		registry.MustRegister(RuntimeActionsCounter)
		registry.MustRegister(RuntimeSuccessActionsCounter)
		registry.MustRegister(RuntimeActionsTimer)
		registry.MustRegister(GRPCMetrics)
	})
}
