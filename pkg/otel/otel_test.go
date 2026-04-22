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
	"testing"

	"go.opentelemetry.io/otel"
)

func TestNewOTelOptions_Defaults(t *testing.T) {
	opts := NewOTelOptions()
	if opts.Mode != OutputModeNoop {
		t.Errorf("expected default mode %q, got %q", OutputModeNoop, opts.Mode)
	}
	if opts.Insecure {
		t.Error("expected Insecure to default to false (secure by default)")
	}
}

func TestValidate_NoopMode(t *testing.T) {
	opts := NewOTelOptions()
	errs := opts.Validate()
	if len(errs) != 0 {
		t.Errorf("noop mode should have no validation errors, got %v", errs)
	}
}

func TestValidate_OTLPModeRequiresEndpoint(t *testing.T) {
	opts := &OTelOptions{Mode: OutputModeOTLP, ServiceName: "test"}
	errs := opts.Validate()
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for missing endpoint, got %d: %v", len(errs), errs)
	}
}

func TestValidate_OTLPModeRequiresServiceName(t *testing.T) {
	opts := &OTelOptions{Mode: OutputModeOTLP, Endpoint: "localhost:4317"}
	errs := opts.Validate()
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for missing service-name, got %d: %v", len(errs), errs)
	}
}

func TestValidate_PrometheusModeRequiresServiceName(t *testing.T) {
	opts := &OTelOptions{Mode: OutputModePrometheus}
	errs := opts.Validate()
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for missing service-name, got %d: %v", len(errs), errs)
	}
}

func TestValidate_UnknownMode(t *testing.T) {
	opts := &OTelOptions{Mode: "invalid"}
	errs := opts.Validate()
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for unknown mode, got %d: %v", len(errs), errs)
	}
}

func TestApply_NoopMode(t *testing.T) {
	opts := &OTelOptions{Mode: OutputModeNoop, ServiceName: "test-noop"}
	providers, err := opts.Apply(context.Background())
	if err != nil {
		t.Fatalf("Apply(noop) failed: %v", err)
	}
	defer func() { _ = providers.Shutdown(context.Background()) }()

	if providers.tracerProvider == nil {
		t.Error("tracerProvider should not be nil in noop mode")
	}
	if providers.meterProvider != nil {
		t.Error("meterProvider should be nil in noop mode")
	}
}

func TestApply_PrometheusMode_NoEndpoint(t *testing.T) {
	opts := &OTelOptions{Mode: OutputModePrometheus, ServiceName: "test-prom"}
	providers, err := opts.Apply(context.Background())
	if err != nil {
		t.Fatalf("Apply(prometheus) failed: %v", err)
	}
	defer func() { _ = providers.Shutdown(context.Background()) }()

	if providers.tracerProvider == nil {
		t.Error("tracerProvider should not be nil")
	}
	if providers.meterProvider == nil {
		t.Error("meterProvider should not be nil in prometheus mode")
	}
}

func TestApply_SetsGlobalProvider(t *testing.T) {
	opts := &OTelOptions{Mode: OutputModeNoop, ServiceName: "test-global"}
	providers, err := opts.Apply(context.Background())
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	defer func() { _ = providers.Shutdown(context.Background()) }()

	tp := otel.GetTracerProvider()
	if tp == nil {
		t.Fatal("global TracerProvider should be set after Apply")
	}
}

func TestProviders_Shutdown(t *testing.T) {
	opts := &OTelOptions{Mode: OutputModePrometheus, ServiceName: "test-shutdown"}
	providers, err := opts.Apply(context.Background())
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}
	if err := providers.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown returned unexpected error: %v", err)
	}
}

func TestInitializer(t *testing.T) {
	opts := &OTelOptions{Mode: OutputModeNoop, ServiceName: "test-initializer"}
	init := opts.Initializer()

	shutdown, err := init(context.Background())
	if err != nil {
		t.Fatalf("Initializer() returned error: %v", err)
	}
	if shutdown == nil {
		t.Fatal("shutdown function should not be nil")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("shutdown returned error: %v", err)
	}
}

func TestApply_UnknownMode(t *testing.T) {
	opts := &OTelOptions{Mode: "bogus", ServiceName: "test"}
	_, err := opts.Apply(context.Background())
	if err == nil {
		t.Fatal("Apply should fail for unknown mode")
	}
}
