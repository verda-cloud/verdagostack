package options

import (
	"time"

	"github.com/spf13/pflag"
)

var _ IOptions = (*SecureServingOptions)(nil)

// SecureServingOptions defines the configuration for secure (TLS) serving.
// CA, Cert, and Key fields can be either file paths or the actual PEM/Base64 content.
type SecureServingOptions struct {
	Addr       string        `json:"addr" mapstructure:"addr"`
	Timeout    time.Duration `json:"timeout" mapstructure:"timeout"`
	TLSOptions `json:",inline" mapstructure:",squash"`
}

// NewSecureServingOptions creates a zero-value instance with defaults.
func NewSecureServingOptions() *SecureServingOptions {
	return &SecureServingOptions{
		Addr:       ":443",
		Timeout:    10 * time.Second,
		TLSOptions: NewTLSOptions(),
	}
}

// Validate verifies the flags passed to SecureServingOptions.
func (o *SecureServingOptions) Validate() []error {
	var errs []error
	errs = append(errs, o.TLSOptions.Validate()...)
	return errs
}

// AddFlags adds flags to the specified FlagSet.
func (o *SecureServingOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.StringVar(&o.Addr, fullPrefix+".addr", o.Addr, "The address the secure server listens on (e.g. :443).")
	fs.DurationVar(&o.Timeout, fullPrefix+".timeout", o.Timeout, "Timeout for TLS connections.")
	o.TLSOptions.AddFlags(fs, fullPrefix)
}
