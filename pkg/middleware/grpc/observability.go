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

// Package grpc provides observability interceptors for gRPC servers.
// It extracts OTel trace context and injects trace metadata into gRPC
// response headers, with structured request logging.
package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	TraceParentHeaderKey = "traceparent"
	TraceIDHeaderKey     = "X-Trace-Id"
	RequestIDHeaderKey   = "X-Request-Id"
	TraceStateHeaderKey  = "tracestate"
)

// TraceInjectionMode defines how trace information is injected into gRPC response metadata.
type TraceInjectionMode int

const (
	InjectW3CTraceContext TraceInjectionMode = iota
	InjectTraceIDOnly
	InjectBoth
	InjectNone
)

// ObservabilityOptions configures the gRPC observability interceptor.
type ObservabilityOptions struct {
	TraceInjectionMode TraceInjectionMode
	CustomTraceHeader  string
}

// Option is a functional option for configuring the interceptor.
type Option func(*ObservabilityOptions)

// WithTraceInjection configures the trace injection mode.
func WithTraceInjection(mode TraceInjectionMode) Option {
	return func(o *ObservabilityOptions) { o.TraceInjectionMode = mode }
}

// WithCustomTraceHeader sets a custom metadata key for the trace ID.
func WithCustomTraceHeader(header string) Option {
	return func(o *ObservabilityOptions) { o.CustomTraceHeader = header }
}

// Observability returns a gRPC UnaryServerInterceptor that injects trace
// metadata into response headers and logs structured request data.
// Assumes an OTel gRPC interceptor (e.g., otelgrpc) has already created spans.
func Observability(opts ...Option) grpc.UnaryServerInterceptor {
	cfg := &ObservabilityOptions{TraceInjectionMode: InjectTraceIDOnly}
	for _, opt := range opts {
		opt(cfg)
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		spanCtx := trace.SpanFromContext(ctx).SpanContext()

		md, _ := metadata.FromIncomingContext(ctx)
		md = injectTraceMetadata(ctx, md, spanCtx, cfg)
		ctx = metadata.NewIncomingContext(ctx, md)

		isDebugLevel := slog.Default().Enabled(ctx, slog.LevelDebug)

		var requestBody string
		if isDebugLevel && req != nil {
			data, _ := json.Marshal(req)
			requestBody = string(data)
		}

		resp, err := handler(ctx, req)

		var responseBody string
		if isDebugLevel && resp != nil {
			data, _ := json.Marshal(resp)
			responseBody = string(data)
		}

		duration := time.Since(start).Seconds()

		var clientIP string
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		status := "OK"
		if err != nil {
			status = "ERROR"
		}

		rpcData := map[string]any{
			"request": map[string]any{
				"method": info.FullMethod,
			},
			"response": map[string]any{
				"status": status,
			},
		}

		if isDebugLevel {
			rpcData["request"].(map[string]any)["body"] = map[string]any{
				"content": requestBody,
				"bytes":   len(requestBody),
			}
			rpcData["response"].(map[string]any)["body"] = map[string]any{
				"content": responseBody,
				"bytes":   len(responseBody),
			}
		}

		logLevel := slog.LevelInfo
		if isDebugLevel {
			logLevel = slog.LevelDebug
		}

		slog.Log(ctx, logLevel, "gRPC request completed",
			"duration_sec", duration,
			"source", map[string]any{"ip": clientIP},
			"rpc", rpcData,
			"trace", map[string]any{"id": spanCtx.TraceID().String()},
			"span", map[string]any{"id": spanCtx.SpanID().String()},
		)

		return resp, err
	}
}

func injectTraceMetadata(ctx context.Context, md metadata.MD, spanCtx trace.SpanContext, cfg *ObservabilityOptions) metadata.MD {
	if !spanCtx.IsValid() {
		return md
	}

	traceID := spanCtx.TraceID().String()
	spanID := spanCtx.SpanID().String()
	traceFlags := "01"
	if !spanCtx.IsSampled() {
		traceFlags = "00"
	}

	switch cfg.TraceInjectionMode {
	case InjectW3CTraceContext:
		md.Set(TraceParentHeaderKey, fmt.Sprintf("00-%s-%s-%s", traceID, spanID, traceFlags))
	case InjectTraceIDOnly:
		header := TraceIDHeaderKey
		if cfg.CustomTraceHeader != "" {
			header = cfg.CustomTraceHeader
		}
		md.Set(header, traceID)
	case InjectBoth:
		md.Set(TraceParentHeaderKey, fmt.Sprintf("00-%s-%s-%s", traceID, spanID, traceFlags))
		header := TraceIDHeaderKey
		if cfg.CustomTraceHeader != "" {
			header = cfg.CustomTraceHeader
		}
		md.Set(header, traceID)
	case InjectNone:
	}

	// Set the trace metadata as gRPC response headers.
	// grpc.SetHeader queues the metadata to be sent with the response.
	_ = grpc.SetHeader(ctx, md)

	return md
}

// Convenience constructors for common configurations.

func ObservabilityWithW3CTraceContext() grpc.UnaryServerInterceptor {
	return Observability(WithTraceInjection(InjectW3CTraceContext))
}

func ObservabilityWithTraceID() grpc.UnaryServerInterceptor {
	return Observability(WithTraceInjection(InjectTraceIDOnly))
}

func ObservabilityWithCustomHeader(headerName string) grpc.UnaryServerInterceptor {
	return Observability(WithTraceInjection(InjectTraceIDOnly), WithCustomTraceHeader(headerName))
}
