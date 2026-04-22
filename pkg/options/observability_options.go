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

package options

import (
	"log/slog"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
)

var _ IOptions = (*ObservabilityOptions)(nil)

// ObservabilityOptions configures the auxiliary observability HTTP server that
// serves health checks, Prometheus metrics, and optionally pprof endpoints.
type ObservabilityOptions struct {
	// BindAddress is the IP address and port to bind the server to.
	BindAddress string `json:"bind-address" mapstructure:"bind-address"`
	// EnablePprof toggles the exposure of runtime profiling data via HTTP.
	EnablePprof bool `json:"enable-pprof" mapstructure:"enable-pprof"`
}

// NewObservabilityOptions returns options with production-ready defaults.
func NewObservabilityOptions() *ObservabilityOptions {
	return &ObservabilityOptions{
		BindAddress: "0.0.0.0:9090",
		EnablePprof: false,
	}
}

// Validate checks the options fields for invalid values.
func (o *ObservabilityOptions) Validate() []error {
	if o == nil {
		return nil
	}
	var errs []error
	if err := ValidateAddress(o.BindAddress); err != nil {
		errs = append(errs, err)
	}
	return errs
}

// AddFlags registers observability flags on the given FlagSet.
func (o *ObservabilityOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.StringVar(&o.BindAddress, fullPrefix+".bind-address", o.BindAddress,
		"Bind address for health checks and metrics (e.g., :9090).")
	fs.BoolVar(&o.EnablePprof, fullPrefix+".enable-pprof", o.EnablePprof,
		"Expose runtime profiling data via HTTP.")
}

// Serve starts the observability HTTP server with default handlers.
// Default handlers: "/healthz" (health check) and "/metrics" (Prometheus).
// This method blocks; call it in a goroutine.
func (o *ObservabilityOptions) Serve() error {
	return o.ServeWithHandlers(nil)
}

// ServeWithHandlers starts the server with default handlers plus any extras.
// Default handlers (/healthz, /metrics) are always registered first;
// extraHandlers can override them if the same path is provided.
func (o *ObservabilityOptions) ServeWithHandlers(extraHandlers map[string]http.Handler) error {
	handlers := map[string]http.Handler{
		"/healthz": o.DefaultHealthzHandler(),
		"/metrics": o.DefaultMetricsHandler(),
	}
	for path, handler := range extraHandlers {
		handlers[path] = handler
	}

	mux := http.NewServeMux()

	for path, handler := range handlers {
		h := handler
		mux.Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.ServeHTTP(w, r)
		}))
		slog.Info("registered observability handler", "path", path)
	}

	if o.EnablePprof {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		slog.Info("pprof profiling enabled", "path_prefix", "/debug/pprof/")
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code": "NotFound", "message": "The requested page was not found."}`))
	})

	server := &http.Server{
		Addr:              o.BindAddress,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	slog.Info("starting observability server", "addr", o.BindAddress)
	return server.ListenAndServe()
}

// DefaultMetricsHandler returns a Prometheus metrics handler that scrapes
// from the default gatherer (where OTel Prometheus reader registers metrics).
func (o *ObservabilityOptions) DefaultMetricsHandler() http.Handler {
	return promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	)
}

// DefaultHealthzHandler returns a simple liveness check handler.
func (o *ObservabilityOptions) DefaultHealthzHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
}
