package metrics

import (
	"sync"

	"github.com/alibaba/pouch/pkg/utils/metrics"
)

func init() {
	// Register prometheus metrics.
	Register()
}

const (
	subsystemPouch = "daemon"
)

var (
	// ImagePullSummary records the summary of pulling image latency.
	ImagePullSummary = metrics.NewLabelSummary(subsystemPouch, "image_pull_latency_microseconds", "Latency in microseconds to pull a image.", "image")

	// ContainerActionsCounter records the number of container operations.
	ContainerActionsCounter = metrics.NewLabelCounter(subsystemPouch, "container_actions_counter", "The number of container operations", "action")

	// ContainerSuccessActionsCounter records the number of container success operations.
	ContainerSuccessActionsCounter = metrics.NewLabelCounter(subsystemPouch, "container_success_actions_counter", "The number of container success operations", "action")

	// ImageActionsCounter records the number of image operations.
	ImageActionsCounter = metrics.NewLabelCounter(subsystemPouch, "image_actions_counter", "The number of image operations", "action")

	// ImageSuccessActionsCounter the number of image success operations.
	ImageSuccessActionsCounter = metrics.NewLabelCounter(subsystemPouch, "image_success_actions_counter", "The number of image success operations", "action")

	// ContainerActionsTimer records the time cost of each container action.
	ContainerActionsTimer = metrics.NewLabelTimer(subsystemPouch, "container_actions", "The number of seconds it takes to process each container action", "action")

	// ImageActionsTimer records the time cost of each image action.
	ImageActionsTimer = metrics.NewLabelTimer(subsystemPouch, "image_actions", "The number of seconds it takes to process each image action", "action")

	// EngineVersion records the version and commit information of the engine process.
	EngineVersion = metrics.NewLabelGauge(subsystemPouch, "engine", "The version and commit information of the engine process", "commit", "version", "kernel")
)

var registerMetrics sync.Once

// Register all metrics.
func Register() {
	// Get a prometheus registry.
	registry := metrics.GetPrometheusRegistry()
	registerMetrics.Do(func() {
		// Register the custom metrics.
		registry.MustRegister(ImagePullSummary)
		registry.MustRegister(EngineVersion)
		registry.MustRegister(ContainerActionsCounter)
		registry.MustRegister(ContainerSuccessActionsCounter)
		registry.MustRegister(ImageActionsCounter)
		registry.MustRegister(ImageSuccessActionsCounter)
		registry.MustRegister(ContainerActionsTimer)
		registry.MustRegister(ImageActionsTimer)
	})
}
