package metrics

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace = "engine"
)

var (
	prometheusRegistry *prometheus.Registry
	prometheusHandler  http.Handler
	registerMetrics    sync.Once
)

func init() {
	prometheusRegistry = prometheus.NewRegistry()
	prometheusHandler = promhttp.HandlerFor(prometheusRegistry, promhttp.HandlerOpts{})
	registerDefaultMetrics(prometheusRegistry)
}

// SinceInMicroseconds gets the time since the specified start in microseconds.
func SinceInMicroseconds(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds() / time.Microsecond.Nanoseconds())
}

// NewLabelSummary return a new SummaryVec
func NewLabelSummary(subsystem, name, help string, labels ...string) *prometheus.SummaryVec {
	return prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        name,
			Help:        help,
			ConstLabels: nil,
		}, labels)
}

// NewLabelCounter return a new CounterVec
func NewLabelCounter(subsystem, name, help string, labels ...string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        fmt.Sprintf("%s_%s", name, total),
			Help:        help,
			ConstLabels: nil,
		}, labels)
}

// NewLabelGauge return a new GaugeVec
func NewLabelGauge(subsystem, name, help string, labels ...string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        fmt.Sprintf("%s_%s", name, Unit("info")),
			Help:        help,
			ConstLabels: nil,
		}, labels)
}

// NewLabelTimer return a new HistogramVec
func NewLabelTimer(subsystem, name, help string, labels ...string) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        fmt.Sprintf("%s_%s", name, seconds),
			Help:        help,
			ConstLabels: nil,
		}, labels)
}

// GetPrometheusRegistry return a registry of Prometheus.
func GetPrometheusRegistry() *prometheus.Registry {
	return prometheusRegistry
}

// GetPrometheusHandler return the prometheus handler.
func GetPrometheusHandler() http.Handler {
	return prometheusHandler
}

func registerDefaultMetrics(registry *prometheus.Registry) {
	//Register the default metrics to the registry in prometheus.
	registerMetrics.Do(func() {
		registry.MustRegister(prometheus.NewProcessCollector(os.Getpid(), ""))
		registry.MustRegister(prometheus.NewGoCollector())
	})
}
