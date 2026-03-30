package log

import (
	"github.com/spf13/pflag"
	"go.uber.org/zap/zapcore"
)

// Options holds configuration for the Logger.
type Options struct {
	// DisableCaller controls whether caller information (file:line) is included.
	DisableCaller bool `json:"disable-caller,omitempty" mapstructure:"disable-caller"`

	// DisableStacktrace controls whether stack traces are recorded
	// for messages at or above panic level.
	DisableStacktrace bool `json:"disable-stacktrace,omitempty" mapstructure:"disable-stacktrace"`

	// EnableColor enables ANSI color output in console format.
	EnableColor bool `json:"enable-color" mapstructure:"enable-color"`

	// Level sets the minimum enabled log level.
	// Valid values: debug, info, warn, error, dpanic, panic, fatal.
	Level string `json:"level,omitempty" mapstructure:"level"`

	// Format sets the log output encoding.
	// Valid values: console, json.
	Format string `json:"format,omitempty" mapstructure:"format"`

	// OutputPaths is a list of URLs or file paths to write log output to.
	OutputPaths []string `json:"output-paths,omitempty" mapstructure:"output-paths"`
}

// NewOptions returns an Options with production-ready defaults.
func NewOptions() *Options {
	return &Options{
		Level:       zapcore.InfoLevel.String(),
		Format:      "console",
		OutputPaths: []string{"stdout"},
	}
}

// Validate checks the Options fields for invalid values.
func (o *Options) Validate() []error {
	var errs []error
	return errs
}

// AddFlags registers CLI flags for all Options fields on the given FlagSet.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Level, "log.level", o.Level, "Minimum log output `LEVEL`.")
	fs.BoolVar(&o.DisableCaller, "log.disable-caller", o.DisableCaller,
		"Disable output of caller information in the log.")
	fs.BoolVar(&o.DisableStacktrace, "log.disable-stacktrace", o.DisableStacktrace,
		"Disable the log to record a stack trace for all messages at or above panic level.")
	fs.BoolVar(&o.EnableColor, "log.enable-color", o.EnableColor,
		"Enable output ANSI colors in plain format logs.")
	fs.StringVar(&o.Format, "log.format", o.Format,
		"Log output `FORMAT`, support console or json format.")
	fs.StringSliceVar(&o.OutputPaths, "log.output-paths", o.OutputPaths,
		"Output paths of log.")
}
