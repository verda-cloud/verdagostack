package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"testing"
	"time"
)

// fakeServer is a minimal Server implementation for testing Serve().
type fakeServer struct {
	runFunc          func(ctx context.Context) error
	gracefulStopFunc func(ctx context.Context) error
}

func (f *fakeServer) Run(ctx context.Context) error {
	if f.runFunc != nil {
		return f.runFunc(ctx)
	}
	<-ctx.Done()
	return http.ErrServerClosed
}

func (f *fakeServer) GracefulStop(ctx context.Context) error {
	if f.gracefulStopFunc != nil {
		return f.gracefulStopFunc(ctx)
	}
	return nil
}

func TestServe_GracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	srv := &fakeServer{
		runFunc: func(ctx context.Context) error {
			<-ctx.Done()
			return http.ErrServerClosed
		},
	}

	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, srv)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Serve returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return within timeout")
	}
}

func TestServe_RunErrorBeforeCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runErr := errors.New("bind: address already in use")
	srv := &fakeServer{
		runFunc: func(ctx context.Context) error {
			return runErr
		},
	}

	err := Serve(ctx, srv)
	if !errors.Is(err, runErr) {
		t.Errorf("expected run error %v, got %v", runErr, err)
	}
}

func TestServe_RunErrorDuringShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	realErr := errors.New("unexpected run failure during shutdown")
	srv := &fakeServer{
		runFunc: func(ctx context.Context) error {
			<-ctx.Done()
			return realErr
		},
	}

	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, srv)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, realErr) {
			t.Errorf("expected run error %v, got %v", realErr, err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return within timeout")
	}
}

func TestServe_GracefulStopError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	shutdownErr := errors.New("shutdown timeout")
	srv := &fakeServer{
		runFunc: func(ctx context.Context) error {
			<-ctx.Done()
			return http.ErrServerClosed
		},
		gracefulStopFunc: func(ctx context.Context) error {
			return shutdownErr
		},
	}

	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, srv)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, shutdownErr) {
			t.Errorf("expected shutdown error %v, got %v", shutdownErr, err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return within timeout")
	}
}

func TestProtocolName(t *testing.T) {
	plain := &http.Server{ReadHeaderTimeout: time.Second} //nolint:gosec // test only
	if name := protocolName(plain); name != "http" {
		t.Errorf("expected 'http', got %q", name)
	}

	secure := &http.Server{TLSConfig: &tls.Config{}, ReadHeaderTimeout: time.Second} //nolint:gosec // test only
	if name := protocolName(secure); name != "https" {
		t.Errorf("expected 'https', got %q", name)
	}
}
