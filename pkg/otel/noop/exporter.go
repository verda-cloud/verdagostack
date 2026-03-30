// Package noop provides a no-op span exporter that discards all spans.
// Used when tracing is disabled (e.g., CLI tools, tests, prometheus-only mode).
package noop

import (
	"context"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var _ sdktrace.SpanExporter = (*Exporter)(nil)

// Exporter is a SpanExporter that silently discards all exported spans.
type Exporter struct{}

// NewExporter returns a new no-op span exporter.
func NewExporter() *Exporter {
	return &Exporter{}
}

// ExportSpans discards all spans.
func (e *Exporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return nil
}

// Shutdown is a no-op.
func (e *Exporter) Shutdown(ctx context.Context) error {
	return nil
}
