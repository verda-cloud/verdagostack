package options

import (
	"github.com/spf13/pflag"
)

var _ IOptions = (*ConsulOptions)(nil)

// ConsulOptions defines options for a Consul client connection.
type ConsulOptions struct {
	Addr   string `json:"addr,omitempty" mapstructure:"addr"`
	Scheme string `json:"scheme,omitempty" mapstructure:"scheme"`
}

// NewConsulOptions creates a ConsulOptions object with default parameters.
func NewConsulOptions() *ConsulOptions {
	return &ConsulOptions{
		Addr:   "127.0.0.1:8500",
		Scheme: "http",
	}
}

// Validate verifies flags passed to ConsulOptions.
func (o *ConsulOptions) Validate() []error {
	return nil
}

// AddFlags adds Consul flags to the specified FlagSet.
func (o *ConsulOptions) AddFlags(fs *pflag.FlagSet, fullPrefix string) {
	fs.StringVar(&o.Addr, fullPrefix+".addr", o.Addr, "Address of the Consul server.")
	fs.StringVar(&o.Scheme, fullPrefix+".scheme", o.Scheme, "URI scheme for the Consul server.")
}
