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

package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// Server defines the interface for all server types.
type Server interface {
	// Run starts the server and blocks until it stops or fails.
	Run(ctx context.Context) error
	// GracefulStop gracefully shuts down the server within the given context deadline.
	GracefulStop(ctx context.Context) error
}

// Serve starts the server in a goroutine and blocks until the context is canceled.
// When the context is done, it gracefully shuts down the server with a 10-second timeout.
// If the server fails before or during shutdown, that error is returned.
func Serve(ctx context.Context, srv Server) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Run(ctx)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	slog.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	shutdownErr := srv.GracefulStop(shutdownCtx)

	// Drain the run goroutine so it doesn't leak; ignore http.ErrServerClosed
	// which is the normal result of a graceful shutdown.
	if runErr := <-errCh; runErr != nil && !errors.Is(runErr, http.ErrServerClosed) {
		return runErr
	}

	if shutdownErr != nil {
		return shutdownErr
	}

	slog.Info("server exited successfully")
	return nil
}

// protocolName returns the protocol name based on the server's TLS configuration.
func protocolName(server *http.Server) string {
	if server.TLSConfig != nil {
		return "https"
	}
	return "http"
}
