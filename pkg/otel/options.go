package otel

import (
	"fmt"

	"github.com/spf13/pflag"
)

// OutputMode controls how traces and metrics are exported.
type OutputMode string

const (
	// OutputModeOTLP exports traces and metrics via OTLP gRPC to a collector.
	OutputModeOTLP OutputMode = "otlp"
	// OutputModePrometheus exports traces via OTLP gRPC (if endpoint set) and
	// metrics via Prometheus scrape endpoint.
	OutputModePrometheus OutputMode = "prometheus"
	// OutputModeNoop disables all telemetry. Suitable for CLI tools and tests.
	OutputModeNoop OutputMode = "noop"
)

// OTelOptions configures the OpenTelemetry providers for traces and metrics.
type OTelOptions struct {
	// Mode selects the export backend: "otlp", "prometheus", or "noop".
	Mode OutputMode `json:"mode" mapstructure:"mode"`
	// Endpoint is the OTLP collector address (e.g., "otel-collector:4317").
	// Required for "otlp" mode, optional for "prometheus" mode (traces only).
	Endpoint string `json:"endpoint" mapstructure:"endpoint"`
	// ServiceName identifies this service in traces and metrics.
	ServiceName string `json:"service-name" mapstructure:"service-name"`
	// Insecure disables TLS for the OTLP gRPC connection.
	Insecure bool `json:"insecure" mapstructure:"insecure"`
}

// NewOTelOptions returns OTelOptions with safe defaults (noop mode).
func NewOTelOptions() *OTelOptions {
	return &OTelOptions{
		Mode:     OutputModeNoop,
		Insecure: true,
	}
}

// Validate checks that the options are consistent.
func (o *OTelOptions) Validate() []error {
	var errs []error

	switch o.Mode {
	case OutputModeOTLP:
		if o.Endpoint == "" {
			errs = append(errs, fmt.Errorf("otel endpoint is required in %q mode", o.Mode))
		}
		if o.ServiceName == "" {
			errs = append(errs, fmt.Errorf("otel service-name is required in %q mode", o.Mode))
		}
	case OutputModePrometheus:
		if o.ServiceName == "" {
			errs = append(errs, fmt.Errorf("otel service-name is required in %q mode", o.Mode))
		}
	case OutputModeNoop:
		// no validation needed
	default:
		errs = append(errs, fmt.Errorf("unknown otel mode %q (valid: otlp, prometheus, noop)", o.Mode))
	}

	return errs
}

// AddFlags registers OTel flags on the given FlagSet.
func (o *OTelOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.StringVar((*string)(&o.Mode), fullPrefix+".mode", string(o.Mode),
		`Telemetry export mode: "otlp", "prometheus", or "noop".`)
	fs.StringVar(&o.Endpoint, fullPrefix+".endpoint", o.Endpoint,
		"OTLP collector gRPC endpoint (e.g., otel-collector:4317).")
	fs.StringVar(&o.ServiceName, fullPrefix+".service-name", o.ServiceName,
		"Service name reported in traces and metrics.")
	fs.BoolVar(&o.Insecure, fullPrefix+".insecure", o.Insecure,
		"Disable TLS for the OTLP gRPC connection.")
}
