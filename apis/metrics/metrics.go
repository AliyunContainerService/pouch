package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	// Register prometheus metrics.
	Register()
}

const (
	namespace = "engine"
	subsystem = "daemon"
)

var (
	// ImagePullSummary records the summary of pulling image latency.
	ImagePullSummary = newLabelSummary("image_pull_latency_microseconds", "Latency in microseconds to pull a image.", "image")

	// ContainerActionsCounter records the number of container operations.
	ContainerActionsCounter = newLabelCounter("container_actions_counter", "The number of container operations", "action")

	// ContainerSuccessActionsCounter records the number of container success operations.
	ContainerSuccessActionsCounter = newLabelCounter("container_success_actions_counter", "The number of container success operations", "action")

	// ImageActionsCounter records the number of image operations.
	ImageActionsCounter = newLabelCounter("image_actions_counter", "The number of image operations", "action")

	// ImageSuccessActionsCounter the number of image success operations.
	ImageSuccessActionsCounter = newLabelCounter("image_success_actions_counter", "The number of image success operations", "action")

	// ContainerActionsTimer records the time cost of each container action.
	ContainerActionsTimer = newLabelTimer("container_actions", "The number of seconds it takes to process each container action", "action")

	// ImageActionsTimer records the time cost of each image action.
	ImageActionsTimer = newLabelTimer("image_actions", "The number of seconds it takes to process each image action", "action")

	// EngineVersion records the version and commit information of the engine process.
	EngineVersion = newLabelGauge("engine", "The version and commit information of the engine process", "commit")
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

// SinceInMicroseconds gets the time since the specified start in microseconds.
func SinceInMicroseconds(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds() / time.Microsecond.Nanoseconds())
}

func newLabelSummary(name, help string, labels ...string) *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        name,
			Help:        help,
			ConstLabels: nil,
		}, labels)
}

func newLabelCounter(name, help string, labels ...string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        fmt.Sprintf("%s_%s", name, total),
			Help:        help,
			ConstLabels: nil,
		}, labels)
}

func newLabelGauge(name, help string, labels ...string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        fmt.Sprintf("%s_%s", name, Unit("info")),
			Help:        help,
			ConstLabels: nil,
		}, labels)
}

func newLabelTimer(name, help string, labels ...string) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        fmt.Sprintf("%s_%s", name, seconds),
			Help:        help,
			ConstLabels: nil,
		}, labels)
}
