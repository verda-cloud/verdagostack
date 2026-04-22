// Copyright 2026 Verda Cloud Oy
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
