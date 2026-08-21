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

// Package otel provides a one-call setup for OpenTelemetry tracing and metrics.
//
// Three output modes are supported:
//   - "otlp": traces + metrics exported via OTLP gRPC
//   - "prometheus": traces via OTLP gRPC (optional), metrics via Prometheus scrape
//   - "noop": all telemetry disabled
//
// Usage:
//
//	providers, err := otelOpts.Apply(ctx)
//	if err != nil { return err }
//	defer providers.Shutdown(context.Background())
package otel

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	noopexporter "github.com/verda-cloud/verdagostack/pkg/otel/noop"
)

// Providers holds the initialized OTel providers and exposes a Shutdown method
// for graceful teardown.
type Providers struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
}

// Shutdown flushes and shuts down all providers. Returns the first error encountered.
func (p *Providers) Shutdown(ctx context.Context) error {
	var firstErr error
	if p.tracerProvider != nil {
		firstErr = p.tracerProvider.Shutdown(ctx)
	}
	if p.meterProvider != nil {
		if err := p.meterProvider.Shutdown(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Apply initializes OTel providers based on the configured mode and registers
// them as the global tracer/meter providers and text map propagator.
func (o *OTelOptions) Apply(ctx context.Context) (*Providers, error) {
	res, err := o.buildResource(ctx)
	if err != nil {
		return nil, fmt.Errorf("otel: build resource: %w", err)
	}

	p := &Providers{}

	switch o.Mode {
	case OutputModeOTLP:
		if err := o.initOTLP(ctx, res, p); err != nil {
			return nil, err
		}
	case OutputModePrometheus:
		if err := o.initPrometheus(ctx, res, p); err != nil {
			return nil, err
		}
	case OutputModeNoop:
		o.initNoop(res, p)
	default:
		return nil, fmt.Errorf("otel: unknown mode %q", o.Mode)
	}

	otel.SetTracerProvider(p.tracerProvider)
	if p.meterProvider != nil {
		otel.SetMeterProvider(p.meterProvider)
	}
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	slog.Info("OpenTelemetry initialized", "mode", o.Mode, "service", o.ServiceName)
	return p, nil
}

func (o *OTelOptions) buildResource(ctx context.Context) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithAttributes(semconv.ServiceName(o.ServiceName)),
		resource.WithProcessRuntimeDescription(),
		resource.WithTelemetrySDK(),
	)
}

func (o *OTelOptions) dialOpts() []grpc.DialOption {
	var opts []grpc.DialOption
	if o.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	return opts
}

// initOTLP sets up both traces and metrics via OTLP gRPC.
func (o *OTelOptions) initOTLP(ctx context.Context, res *resource.Resource, p *Providers) error {
	conn, err := grpc.NewClient(o.Endpoint, o.dialOpts()...)
	if err != nil {
		return fmt.Errorf("otel: dial %s: %w", o.Endpoint, err)
	}

	traceExp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return fmt.Errorf("otel: create trace exporter: %w", err)
	}

	metricExp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return fmt.Errorf("otel: create metric exporter: %w", err)
	}

	p.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExp),
	)

	p.meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)),
	)

	return nil
}

// initPrometheus sets up traces via OTLP gRPC (if endpoint provided) or noop,
// and metrics via Prometheus scrape reader registered with DefaultGatherer.
func (o *OTelOptions) initPrometheus(ctx context.Context, res *resource.Resource, p *Providers) error {
	if o.Endpoint != "" {
		conn, err := grpc.NewClient(o.Endpoint, o.dialOpts()...)
		if err != nil {
			return fmt.Errorf("otel: dial %s: %w", o.Endpoint, err)
		}
		traceExp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			return fmt.Errorf("otel: create trace exporter: %w", err)
		}
		p.tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithBatcher(traceExp),
		)
	} else {
		p.tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithBatcher(noopexporter.NewExporter()),
		)
	}

	promReader, err := prometheus.New()
	if err != nil {
		return fmt.Errorf("otel: create prometheus reader: %w", err)
	}

	p.meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(promReader),
	)

	return nil
}

// initNoop disables all telemetry.
func (o *OTelOptions) initNoop(res *resource.Resource, p *Providers) {
	p.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(noopexporter.NewExporter()),
	)
}

// Initializer returns a function suitable for app.WithOTel. It calls Apply
// on first invocation and returns the Shutdown function for deferred cleanup.
func (o *OTelOptions) Initializer() func(ctx context.Context) (func(context.Context) error, error) {
	return func(ctx context.Context) (func(context.Context) error, error) {
		p, err := o.Apply(ctx)
		if err != nil {
			return nil, err
		}
		return p.Shutdown, nil
	}
}
