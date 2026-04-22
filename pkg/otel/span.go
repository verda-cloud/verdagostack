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

package otel

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// RecordSpanError records an error on the span with optional attributes and
// sets the span status to Error.
func RecordSpanError(ctx context.Context, span trace.Span, err error, attrs ...attribute.KeyValue) {
	if len(attrs) > 0 {
		span.RecordError(err, trace.WithAttributes(attrs...))
	} else {
		span.RecordError(err)
	}
	span.SetStatus(codes.Error, err.Error())
}

// RecordSpanErrorWithLog records an error on the span, sets the status, and
// emits a structured slog.Error with the given message and attributes.
func RecordSpanErrorWithLog(ctx context.Context, span trace.Span, err error, message string, attrs ...attribute.KeyValue) {
	if len(attrs) > 0 {
		span.RecordError(err, trace.WithAttributes(attrs...))
	} else {
		span.RecordError(err)
	}
	span.SetStatus(codes.Error, message)

	logAttrs := []any{"error", err.Error()}
	for _, attr := range attrs {
		logAttrs = append(logAttrs, string(attr.Key), attr.Value.AsInterface())
	}
	slog.ErrorContext(ctx, message, logAttrs...)
}

// LogSpanError is a convenience wrapper that uses the error message as the log message.
func LogSpanError(ctx context.Context, span trace.Span, err error, attrs ...attribute.KeyValue) {
	RecordSpanErrorWithLog(ctx, span, err, err.Error(), attrs...)
}

// RecordSpanErrorf records a span error with a formatted message.
func RecordSpanErrorf(ctx context.Context, span trace.Span, err error, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	span.RecordError(err)
	span.SetStatus(codes.Error, message)
	slog.ErrorContext(ctx, message, "error", err.Error())
}

// RecordSpanErrorfWithAttrs records a span error with a formatted message and
// additional trace attributes.
func RecordSpanErrorfWithAttrs(ctx context.Context, span trace.Span, err error, format string, args []interface{}, attrs ...attribute.KeyValue) {
	message := fmt.Sprintf(format, args...)

	if len(attrs) > 0 {
		span.RecordError(err, trace.WithAttributes(attrs...))
	} else {
		span.RecordError(err)
	}
	span.SetStatus(codes.Error, message)

	logAttrs := []any{"error", err.Error()}
	for _, attr := range attrs {
		logAttrs = append(logAttrs, string(attr.Key), attr.Value.AsInterface())
	}
	slog.ErrorContext(ctx, message, logAttrs...)
}
