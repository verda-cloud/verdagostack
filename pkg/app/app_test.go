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

package app

import (
	"context"
	"errors"
	"testing"
)

func TestNewApp_Defaults(t *testing.T) {
	app := NewApp("test-app", "A test application")
	if app.name != "test-app" {
		t.Errorf("expected name 'test-app', got %q", app.name)
	}
	if app.cmd == nil {
		t.Fatal("cmd should not be nil")
	}
	if app.cmd.Use != "test-app" {
		t.Errorf("expected Use 'test-app', got %q", app.cmd.Use)
	}
}

func TestNewApp_WithRunFunc_ReceivesContext(t *testing.T) {
	var gotCtx context.Context

	app := NewApp("ctx-test", "test context propagation",
		WithRunFunc(func(ctx context.Context) error {
			gotCtx = ctx
			return nil
		}),
		WithNoConfig(),
		WithSilence(),
	)

	app.cmd.SetArgs([]string{})
	if err := app.cmd.Execute(); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if gotCtx == nil {
		t.Fatal("RunFunc was not called or received nil context")
	}
}

func TestNewApp_WithRunFunc_ErrorPropagated(t *testing.T) {
	expectedErr := errors.New("startup failure")

	app := NewApp("err-test", "test error propagation",
		WithRunFunc(func(ctx context.Context) error {
			return expectedErr
		}),
		WithNoConfig(),
		WithSilence(),
	)

	app.cmd.SetArgs([]string{})
	err := app.cmd.Execute()
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestNewApp_WithSilence(t *testing.T) {
	app := NewApp("silent", "silent app", WithSilence(), WithNoConfig())
	if !app.silence {
		t.Error("expected silence to be true")
	}
}

func TestNewApp_WithNoConfig(t *testing.T) {
	app := NewApp("no-cfg", "no config app", WithNoConfig())
	if !app.noConfig {
		t.Error("expected noConfig to be true")
	}
}

func TestNewApp_WithDescription(t *testing.T) {
	app := NewApp("desc", "short", WithDescription("A long description"))
	if app.description != "A long description" {
		t.Errorf("expected long description, got %q", app.description)
	}
}

func TestNewApp_WithDefaultValidArgs(t *testing.T) {
	app := NewApp("strict", "strict args",
		WithDefaultValidArgs(),
		WithNoConfig(),
		WithSilence(),
		WithRunFunc(func(ctx context.Context) error { return nil }),
	)

	app.cmd.SetArgs([]string{"unexpected-arg"})
	err := app.cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unexpected positional arg")
	}
}

func TestNewApp_Command(t *testing.T) {
	app := NewApp("cmd-test", "command access")
	if app.Command() == nil {
		t.Fatal("Command() should not return nil")
	}
	if app.Command() != app.cmd {
		t.Error("Command() should return the internal cobra command")
	}
}

func TestFormatBaseName(t *testing.T) {
	name := formatBaseName("MyApp")
	if name != "MyApp" {
		t.Errorf("expected 'MyApp' on non-windows, got %q", name)
	}
}

type testOptions struct {
	Name    string `mapstructure:"name"`
	Verbose bool   `mapstructure:"verbose"`
}

func TestNewApp_OptionsValidation(t *testing.T) {
	validationErr := errors.New("name is required")

	type validatable struct {
		testOptions
	}

	opts := &struct {
		validatable
		validateFunc func() error
	}{
		validateFunc: func() error { return validationErr },
	}

	_ = opts

	app := NewApp("validate-test", "opts validation",
		WithOptions(&testOptions{}),
		WithNoConfig(),
		WithSilence(),
		WithRunFunc(func(ctx context.Context) error { return nil }),
	)

	app.cmd.SetArgs([]string{})
	err := app.cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewApp_HealthCheckFunc(t *testing.T) {
	called := false
	app := NewApp("health-test", "health check test",
		WithNoConfig(),
		WithSilence(),
		WithHealthCheckFunc(func() error {
			called = true
			return nil
		}),
		WithRunFunc(func(ctx context.Context) error { return nil }),
	)

	app.cmd.SetArgs([]string{})
	if err := app.cmd.Execute(); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}
	if !called {
		t.Error("health check function was not called")
	}
}

func TestNewApp_HealthCheckFunc_Error(t *testing.T) {
	healthErr := errors.New("health check setup failed")
	app := NewApp("health-err", "health error test",
		WithNoConfig(),
		WithSilence(),
		WithHealthCheckFunc(func() error {
			return healthErr
		}),
		WithRunFunc(func(ctx context.Context) error { return nil }),
	)

	app.cmd.SetArgs([]string{})
	err := app.cmd.Execute()
	if !errors.Is(err, healthErr) {
		t.Errorf("expected health error %v, got %v", healthErr, err)
	}
}

func TestNewApp_WithOTel_InitCalledBeforeRun(t *testing.T) {
	var order []string

	initFn := func(ctx context.Context) (func(context.Context) error, error) {
		order = append(order, "init")
		return func(ctx context.Context) error {
			order = append(order, "shutdown")
			return nil
		}, nil
	}

	app := NewApp("otel-test", "otel lifecycle test",
		WithNoConfig(),
		WithSilence(),
		WithOTel(initFn),
		WithRunFunc(func(ctx context.Context) error {
			order = append(order, "run")
			return nil
		}),
	)

	app.cmd.SetArgs([]string{})
	if err := app.cmd.Execute(); err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if len(order) != 3 {
		t.Fatalf("expected 3 lifecycle events, got %d: %v", len(order), order)
	}
	if order[0] != "init" {
		t.Errorf("expected init first, got %q", order[0])
	}
	if order[1] != "run" {
		t.Errorf("expected run second, got %q", order[1])
	}
	if order[2] != "shutdown" {
		t.Errorf("expected shutdown third, got %q", order[2])
	}
}

func TestNewApp_WithOTel_InitError(t *testing.T) {
	initErr := errors.New("otel init failed")

	app := NewApp("otel-err", "otel error test",
		WithNoConfig(),
		WithSilence(),
		WithOTel(func(ctx context.Context) (func(context.Context) error, error) {
			return nil, initErr
		}),
		WithRunFunc(func(ctx context.Context) error {
			t.Fatal("RunFunc should not be called when OTel init fails")
			return nil
		}),
	)

	app.cmd.SetArgs([]string{})
	err := app.cmd.Execute()
	if !errors.Is(err, initErr) {
		t.Errorf("expected otel init error %v, got %v", initErr, err)
	}
}
