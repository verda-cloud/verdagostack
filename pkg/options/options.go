package options

import "github.com/spf13/pflag"

// IOptions defines methods to implement generic options.
type IOptions interface {
	// Validate validates all the required options.
	Validate() []error

	// AddFlags registers all option fields as command line flags on the given FlagSet,
	// using the provided fullPrefix directly.
	//
	// Example:
	//
	//	o.AddFlags(fs, "server.http")  // --server.http.addr, --server.http.timeout, etc.
	AddFlags(fs *pflag.FlagSet, fullPrefix string)
}
