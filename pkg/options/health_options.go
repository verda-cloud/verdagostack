package options

import (
	"log/slog"
	"net/http"
	"net/http/pprof"

	"github.com/spf13/pflag"
)

var _ IOptions = (*HealthOptions)(nil)

// HealthOptions defines options for the health check and pprof HTTP server.
type HealthOptions struct {
	HTTPProfile        bool   `json:"enable-http-profiler" mapstructure:"enable-http-profiler"`
	HealthCheckPath    string `json:"check-path" mapstructure:"check-path"`
	HealthCheckAddress string `json:"check-address" mapstructure:"check-address"`
}

// NewHealthOptions creates a HealthOptions object with default parameters.
func NewHealthOptions() *HealthOptions {
	return &HealthOptions{
		HTTPProfile:        false,
		HealthCheckPath:    "/healthz",
		HealthCheckAddress: "0.0.0.0:20250",
	}
}

// Validate verifies flags passed to HealthOptions.
func (o *HealthOptions) Validate() []error {
	return nil
}

// AddFlags adds health check flags to the specified FlagSet.
func (o *HealthOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.BoolVar(&o.HTTPProfile, fullPrefix+".enable-http-profiler", o.HTTPProfile, "Expose runtime profiling data via HTTP.")
	fs.StringVar(&o.HealthCheckPath, fullPrefix+".check-path", o.HealthCheckPath, "Specifies liveness health check request path.")
	fs.StringVar(&o.HealthCheckAddress, fullPrefix+".check-address", o.HealthCheckAddress, "Specifies liveness health check bind address.")
}

// Serve starts a health check HTTP server. If HTTPProfile is enabled, pprof
// endpoints are also registered. This method blocks; call it in a goroutine.
func (o *HealthOptions) Serve() {
	mux := http.NewServeMux()

	mux.HandleFunc(o.HealthCheckPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`)) //nolint:errcheck
	})

	if o.HTTPProfile {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	slog.Info("Starting health check server", "path", o.HealthCheckPath, "addr", o.HealthCheckAddress)
	if err := http.ListenAndServe(o.HealthCheckAddress, mux); err != nil { //nolint:gosec
		slog.Error("Error serving health check endpoint", "error", err)
	}
}
