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
