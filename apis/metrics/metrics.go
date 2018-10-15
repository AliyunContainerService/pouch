package metrics

import (
	"sync"

	util_metrics "github.com/alibaba/pouch/pkg/utils/metrics"

	"github.com/prometheus/client_golang/prometheus"
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
	ImagePullSummary = util_metrics.NewLabelSummary(subsystemPouch, "image_pull_latency_microseconds", "Latency in microseconds to pull a image.", "image")

	// ContainerActionsCounter records the number of container operations.
	ContainerActionsCounter = util_metrics.NewLabelCounter(subsystemPouch, "container_actions_counter", "The number of container operations", "action")

	// ContainerSuccessActionsCounter records the number of container success operations.
	ContainerSuccessActionsCounter = util_metrics.NewLabelCounter(subsystemPouch, "container_success_actions_counter", "The number of container success operations", "action")

	// ImageActionsCounter records the number of image operations.
	ImageActionsCounter = util_metrics.NewLabelCounter(subsystemPouch, "image_actions_counter", "The number of image operations", "action")

	// ImageSuccessActionsCounter the number of image success operations.
	ImageSuccessActionsCounter = util_metrics.NewLabelCounter(subsystemPouch, "image_success_actions_counter", "The number of image success operations", "action")

	// ContainerActionsTimer records the time cost of each container action.
	ContainerActionsTimer = util_metrics.NewLabelTimer(subsystemPouch, "container_actions", "The number of seconds it takes to process each container action", "action")

	// ImageActionsTimer records the time cost of each image action.
	ImageActionsTimer = util_metrics.NewLabelTimer(subsystemPouch, "image_actions", "The number of seconds it takes to process each image action", "action")

	// EngineVersion records the version and commit information of the engine process.
	EngineVersion = util_metrics.NewLabelGauge(subsystemPouch, "engine", "The version and commit information of the engine process", "commit")
)

var registerMetrics sync.Once

// Register all metrics.
func Register() {
	// Register the metrics.
	registerMetrics.Do(func() {
		prometheus.MustRegister(ImagePullSummary)
		prometheus.MustRegister(EngineVersion)
		prometheus.MustRegister(ContainerActionsCounter)
		prometheus.MustRegister(ContainerSuccessActionsCounter)
		prometheus.MustRegister(ImageActionsCounter)
		prometheus.MustRegister(ImageSuccessActionsCounter)
		prometheus.MustRegister(ContainerActionsTimer)
		prometheus.MustRegister(ImageActionsTimer)
	})
}
