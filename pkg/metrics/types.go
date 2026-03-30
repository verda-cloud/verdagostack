package metrics

import "github.com/prometheus/client_golang/prometheus"

// MetricType identifies the kind of Prometheus metric to create.
type MetricType string

const (
	CounterType   MetricType = "counter"
	GaugeType     MetricType = "gauge"
	HistogramType MetricType = "histogram"
	SummaryType   MetricType = "summary"
)

// Convenience aliases so callers don't need to import prometheus directly.
type (
	CounterVec   = prometheus.CounterVec
	GaugeVec     = prometheus.GaugeVec
	HistogramVec = prometheus.HistogramVec
	SummaryVec   = prometheus.SummaryVec
)

// MetricOpts describes a single metric to register.
type MetricOpts struct {
	Name    string
	Help    string
	Type    MetricType
	Labels  []string  // Label keys this metric supports.
	Buckets []float64 // Histogram bucket boundaries (ignored for other types).
}
