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

// Package metrics provides a shared Prometheus metrics library for verdagostack
// applications. Metrics registered here appear on the /metrics endpoint served
// by pkg/options.ObservabilityOptions (which reads from prometheus.DefaultGatherer).
//
// By default, metrics are registered with prometheus.DefaultRegisterer so they
// are automatically unified with OTel-generated metrics. Use WithRegistry to
// isolate metrics in tests or multi-tenant scenarios.
//
// Usage:
//
//	m := metrics.New(
//	    metrics.WithNamespace("myapp"),
//	    metrics.WithSubsystem("api"),
//	)
//	m.MustRegister(metrics.MetricOpts{
//	    Name:   "requests_total",
//	    Help:   "Total HTTP requests",
//	    Type:   metrics.CounterType,
//	    Labels: []string{"method", "status"},
//	})
//
//	// At runtime:
//	m.UpdateMetric("requests_total", 1, []string{"GET", "200"})
//
//	// Or typed:
//	counter, _ := m.GetCounter("requests_total")
//	counter.WithLabelValues("GET", "200").Inc()
package metrics

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics manages a set of named Prometheus collectors.
type Metrics struct {
	mu         sync.RWMutex
	namespace  string
	subsystem  string
	registerer prometheus.Registerer
	gatherer   prometheus.Gatherer
	collectors map[string]prometheus.Collector
}

// Option configures a Metrics instance.
type Option func(*Metrics)

// WithNamespace sets the Prometheus namespace applied to all registered metrics.
func WithNamespace(ns string) Option {
	return func(m *Metrics) { m.namespace = ns }
}

// WithSubsystem sets the Prometheus subsystem applied to all registered metrics.
func WithSubsystem(sub string) Option {
	return func(m *Metrics) { m.subsystem = sub }
}

// WithRegistry uses a custom prometheus.Registry instead of the default
// registerer/gatherer. Useful for testing or multi-tenant isolation.
func WithRegistry(reg *prometheus.Registry) Option {
	return func(m *Metrics) {
		m.registerer = reg
		m.gatherer = reg
	}
}

// New creates a Metrics instance. Without WithRegistry, metrics are registered
// with prometheus.DefaultRegisterer and gathered from prometheus.DefaultGatherer,
// so they automatically appear on the ObservabilityOptions /metrics endpoint.
func New(opts ...Option) *Metrics {
	m := &Metrics{
		registerer: prometheus.DefaultRegisterer,
		gatherer:   prometheus.DefaultGatherer,
		collectors: make(map[string]prometheus.Collector),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// Register creates and registers a metric described by opts.
// Returns an error if the name is already taken or the collector cannot be registered.
func (m *Metrics) Register(opts MetricOpts) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.collectors[opts.Name]; exists {
		return fmt.Errorf("metrics: %q already registered", opts.Name)
	}

	collector, err := m.createCollector(opts)
	if err != nil {
		return err
	}

	if err := m.registerer.Register(collector); err != nil {
		return fmt.Errorf("metrics: register %q: %w", opts.Name, err)
	}

	m.collectors[opts.Name] = collector
	return nil
}

// MustRegister is like Register but panics on error. Intended for use during
// application initialization where a missing metric is a programming error.
func (m *Metrics) MustRegister(opts MetricOpts) {
	if err := m.Register(opts); err != nil {
		panic(err)
	}
}

// RegisterAll registers multiple metrics. Stops on the first error.
func (m *Metrics) RegisterAll(all []MetricOpts) error {
	for i := range all {
		if err := m.Register(all[i]); err != nil {
			return err
		}
	}
	return nil
}

// UpdateMetric updates a metric by name. The behaviour depends on the type:
//   - Counter: adds value
//   - Gauge: sets value
//   - Histogram: observes value
//   - Summary: observes value
func (m *Metrics) UpdateMetric(name string, value float64, labelValues []string) error {
	m.mu.RLock()
	collector, ok := m.collectors[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("metrics: %q not found", name)
	}

	switch c := collector.(type) {
	case *prometheus.CounterVec:
		c.WithLabelValues(labelValues...).Add(value)
	case *prometheus.GaugeVec:
		c.WithLabelValues(labelValues...).Set(value)
	case *prometheus.HistogramVec:
		c.WithLabelValues(labelValues...).Observe(value)
	case *prometheus.SummaryVec:
		c.WithLabelValues(labelValues...).Observe(value)
	default:
		return fmt.Errorf("metrics: unsupported collector type for %q", name)
	}
	return nil
}

// GetCounter returns the CounterVec registered under name.
func (m *Metrics) GetCounter(name string) (*prometheus.CounterVec, error) {
	c, err := m.get(name)
	if err != nil {
		return nil, err
	}
	cv, ok := c.(*prometheus.CounterVec)
	if !ok {
		return nil, fmt.Errorf("metrics: %q is not a counter", name)
	}
	return cv, nil
}

// GetGauge returns the GaugeVec registered under name.
func (m *Metrics) GetGauge(name string) (*prometheus.GaugeVec, error) {
	c, err := m.get(name)
	if err != nil {
		return nil, err
	}
	gv, ok := c.(*prometheus.GaugeVec)
	if !ok {
		return nil, fmt.Errorf("metrics: %q is not a gauge", name)
	}
	return gv, nil
}

// GetHistogram returns the HistogramVec registered under name.
func (m *Metrics) GetHistogram(name string) (*prometheus.HistogramVec, error) {
	c, err := m.get(name)
	if err != nil {
		return nil, err
	}
	hv, ok := c.(*prometheus.HistogramVec)
	if !ok {
		return nil, fmt.Errorf("metrics: %q is not a histogram", name)
	}
	return hv, nil
}

// GetSummary returns the SummaryVec registered under name.
func (m *Metrics) GetSummary(name string) (*prometheus.SummaryVec, error) {
	c, err := m.get(name)
	if err != nil {
		return nil, err
	}
	sv, ok := c.(*prometheus.SummaryVec)
	if !ok {
		return nil, fmt.Errorf("metrics: %q is not a summary", name)
	}
	return sv, nil
}

// GetCollector returns the raw prometheus.Collector registered under name.
func (m *Metrics) GetCollector(name string) (prometheus.Collector, error) {
	return m.get(name)
}

// Handler returns an http.Handler that serves the metrics from this instance's
// gatherer. When using the default registry this is equivalent to
// promhttp.Handler(); when using a custom registry it scopes to that registry.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.gatherer, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

func (m *Metrics) get(name string) (prometheus.Collector, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.collectors[name]
	if !ok {
		return nil, fmt.Errorf("metrics: %q not found", name)
	}
	return c, nil
}

func (m *Metrics) createCollector(opts MetricOpts) (prometheus.Collector, error) {
	switch opts.Type {
	case CounterType:
		return prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: m.namespace,
			Subsystem: m.subsystem,
			Name:      opts.Name,
			Help:      opts.Help,
		}, opts.Labels), nil

	case GaugeType:
		return prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: m.namespace,
			Subsystem: m.subsystem,
			Name:      opts.Name,
			Help:      opts.Help,
		}, opts.Labels), nil

	case HistogramType:
		hopts := prometheus.HistogramOpts{
			Namespace: m.namespace,
			Subsystem: m.subsystem,
			Name:      opts.Name,
			Help:      opts.Help,
		}
		if len(opts.Buckets) > 0 {
			hopts.Buckets = opts.Buckets
		}
		return prometheus.NewHistogramVec(hopts, opts.Labels), nil

	case SummaryType:
		return prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace: m.namespace,
			Subsystem: m.subsystem,
			Name:      opts.Name,
			Help:      opts.Help,
		}, opts.Labels), nil

	default:
		return nil, fmt.Errorf("metrics: unsupported type %q", opts.Type)
	}
}
