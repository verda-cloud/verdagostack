package options

import (
	"time"

	"github.com/spf13/pflag"
)

var _ IOptions = (*InsecureServingOptions)(nil)

// InsecureServingOptions contains configuration items related to HTTP server startup.
type InsecureServingOptions struct {
	Addr    string        `json:"addr" mapstructure:"addr"`
	Timeout time.Duration `json:"timeout" mapstructure:"timeout"`
}

// NewInsecureServingOptions creates an InsecureServingOptions object with default parameters.
func NewInsecureServingOptions() *InsecureServingOptions {
	return &InsecureServingOptions{
		Addr:    ":8080",
		Timeout: 30 * time.Second,
	}
}

// Validate verifies the flags passed to InsecureServingOptions.
func (o *InsecureServingOptions) Validate() []error {
	if o == nil {
		return nil
	}

	var errors []error

	if err := ValidateAddress(o.Addr); err != nil {
		errors = append(errors, err)
	}

	return errors
}

// AddFlags registers HTTP server related flags to the specified FlagSet.
func (o *InsecureServingOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.StringVar(&o.Addr, fullPrefix+".addr", o.Addr,
		"Listen address for the HTTP server (e.g., :8080, 0.0.0.0:8443).")
	fs.DurationVar(&o.Timeout, fullPrefix+".timeout", o.Timeout,
		"Timeout for incoming HTTP connections.")
}

// Complete fills in any fields not set that are required to have valid data.
func (o *InsecureServingOptions) Complete() error {
	return nil
}
