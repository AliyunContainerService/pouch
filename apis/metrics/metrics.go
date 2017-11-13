package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	// Register prometheus metrics.
	Register()
}

const (
	pouchSubsystem = "pouch"
)

var (
	// ImagePullSummary records the summary of pulling image latency.
	ImagePullSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Subsystem: pouchSubsystem,
			Name:      "image_pull_latency_microseconds",
			Help:      "Latency in microseconds to pull a image.",
		},
		[]string{"image"},
	)
)

var registerMetrics sync.Once

// Register all metrics.
func Register() {
	// Register the metrics.
	registerMetrics.Do(func() {
		prometheus.MustRegister(ImagePullSummary)
	})
}

// SinceInMicroseconds gets the time since the specified start in microseconds.
func SinceInMicroseconds(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds() / time.Microsecond.Nanoseconds())
}
