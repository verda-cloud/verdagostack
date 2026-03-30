package options

import (
	"time"

	"github.com/spf13/pflag"
)

var _ IOptions = (*GRPCOptions)(nil)

// GRPCOptions defines configuration for a gRPC server.
type GRPCOptions struct {
	Network string        `json:"network" mapstructure:"network"`
	Addr    string        `json:"addr" mapstructure:"addr"`
	Timeout time.Duration `json:"timeout" mapstructure:"timeout"`
}

// NewGRPCOptions creates a GRPCOptions object with default parameters.
func NewGRPCOptions() *GRPCOptions {
	return &GRPCOptions{
		Network: "tcp",
		Addr:    "0.0.0.0:39090",
		Timeout: 30 * time.Second,
	}
}

// Validate verifies the flags passed to GRPCOptions.
func (o *GRPCOptions) Validate() []error {
	var errors []error

	if err := ValidateAddress(o.Addr); err != nil {
		errors = append(errors, err)
	}

	return errors
}

// AddFlags adds gRPC server flags to the specified FlagSet.
func (o *GRPCOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.StringVar(&o.Network, fullPrefix+".network", o.Network, "Specify the network for the gRPC server.")
	fs.StringVar(&o.Addr, fullPrefix+".addr", o.Addr, "Specify the gRPC server bind address and port.")
	fs.DurationVar(&o.Timeout, fullPrefix+".timeout", o.Timeout, "Timeout for server connections.")
}
