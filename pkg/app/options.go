package app

import "github.com/spf13/pflag"

// OptionsValidator provides methods to complete and validate options.
type OptionsValidator interface {
	Complete() error
	Validate() error
}

// NamedFlagSetOptions provides access to grouped flag sets and validation.
// Options structs implement this to organize their flags into named sections.
type NamedFlagSetOptions interface {
	Flags() NamedFlagSets
	OptionsValidator
}

// FlagSetOptions defines options that add themselves to a single flat FlagSet.
type FlagSetOptions interface {
	AddFlags(fs *pflag.FlagSet)
	OptionsValidator
}
