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

	"golang.org/x/sync/errgroup"

	genericoptions "github.com/verda-cloud/verdagostack/pkg/options"
)

// HTTPServer represents an HTTP server that supports both HTTP and HTTPS simultaneously.
type HTTPServer struct {
	insecureServer *http.Server
	secureServer   *http.Server
}

// NewHTTPServer creates a new instance of HTTPServer.
func NewHTTPServer(
	insecureOptions *genericoptions.InsecureServingOptions,
	secureOptions *genericoptions.SecureServingOptions,
	handler http.Handler,
) *HTTPServer {
	s := &HTTPServer{}

	if insecureOptions != nil && insecureOptions.Addr != "" {
		s.insecureServer = &http.Server{
			Addr:              insecureOptions.Addr,
			Handler:           handler,
			ReadHeaderTimeout: insecureOptions.Timeout,
			ReadTimeout:       insecureOptions.Timeout,
			WriteTimeout:      insecureOptions.Timeout,
			IdleTimeout:       3 * insecureOptions.Timeout,
		}
	}

	if secureOptions != nil && secureOptions.Enabled {
		tlsConfig := secureOptions.MustTLSConfig()
		s.secureServer = &http.Server{
			Addr:              secureOptions.Addr,
			Handler:           handler,
			TLSConfig:         tlsConfig,
			ReadHeaderTimeout: secureOptions.Timeout,
			ReadTimeout:       secureOptions.Timeout,
			WriteTimeout:      secureOptions.Timeout,
			IdleTimeout:       3 * secureOptions.Timeout,
		}
	}

	return s
}

// Run starts all configured HTTP/HTTPS servers and blocks until one fails
// or the context is canceled.
func (s *HTTPServer) Run(ctx context.Context) error {
	eg, _ := errgroup.WithContext(ctx)

	if s.insecureServer != nil {
		slog.Info("starting insecure HTTP server", "addr", s.insecureServer.Addr)
		eg.Go(func() error {
			if err := s.insecureServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("failed to start insecure HTTP server", "error", err)
				return err
			}
			return nil
		})
	}

	if s.secureServer != nil {
		slog.Info("starting secure HTTPS server", "addr", s.secureServer.Addr)
		eg.Go(func() error {
			// TLSConfig already contains the certificates, so empty strings are passed here.
			if err := s.secureServer.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("failed to start secure HTTPS server", "error", err)
				return err
			}
			return nil
		})
	}

	return eg.Wait()
}

// GracefulStop gracefully shuts down all HTTP/HTTPS servers.
func (s *HTTPServer) GracefulStop(ctx context.Context) error {
	slog.Info("gracefully stopping HTTP(s) servers")

	eg, ctx := errgroup.WithContext(ctx)

	if s.insecureServer != nil {
		eg.Go(func() error {
			return s.insecureServer.Shutdown(ctx)
		})
	}

	if s.secureServer != nil {
		eg.Go(func() error {
			return s.secureServer.Shutdown(ctx)
		})
	}

	if err := eg.Wait(); err != nil {
		slog.Error("one or more servers failed to shutdown gracefully", "error", err)
		return err
	}

	slog.Info("all HTTP(s) servers stopped successfully")
	return nil
}
