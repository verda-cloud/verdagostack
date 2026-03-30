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
