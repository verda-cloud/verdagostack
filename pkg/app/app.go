package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/verda-cloud/verdagostack/pkg/log"
	"github.com/verda-cloud/verdagostack/pkg/options"
	"github.com/verda-cloud/verdagostack/pkg/version"
)

// App is the main structure of a CLI application.
// Create one with NewApp().
type App struct {
	name        string
	shortDesc   string
	description string
	run         RunFunc
	cmd         *cobra.Command
	args        cobra.PositionalArgs

	healthCheckFunc HealthCheckFunc
	otelInit        OTelInitFunc

	options  any
	silence  bool
	noConfig bool
	watch    bool
}

// OTelInitFunc initializes OTel providers and returns a shutdown function.
// Defined as a function type so pkg/app doesn't import pkg/otel directly.
// Use otel.OTelOptions.Initializer() to obtain one.
type OTelInitFunc func(ctx context.Context) (shutdown func(ctx context.Context) error, err error)

// RunFunc defines the application's startup callback function.
// The context is cancelled on SIGINT/SIGTERM, enabling graceful shutdown.
type RunFunc func(ctx context.Context) error

// HealthCheckFunc defines the health check function for the application.
type HealthCheckFunc func() error

// Option configures an App.
type Option func(*App)

// WithOptions sets the options struct that will be populated from flags,
// environment variables, and configuration files.
func WithOptions(opts any) Option {
	return func(app *App) {
		app.options = opts
	}
}

// WithRunFunc sets the function called after configuration is loaded and validated.
func WithRunFunc(run RunFunc) Option {
	return func(app *App) {
		app.run = run
	}
}

// WithDescription sets the long description shown in help output.
func WithDescription(desc string) Option {
	return func(app *App) {
		app.description = desc
	}
}

// WithSilence disables startup banners and config printing.
func WithSilence() Option {
	return func(app *App) {
		app.silence = true
	}
}

// WithNoConfig disables the --config flag.
func WithNoConfig() Option {
	return func(app *App) {
		app.noConfig = true
	}
}

// WithValidArgs sets the positional argument validation function.
func WithValidArgs(args cobra.PositionalArgs) Option {
	return func(app *App) {
		app.args = args
	}
}

// WithDefaultValidArgs sets the default validation (no positional args allowed).
func WithDefaultValidArgs() Option {
	return func(app *App) {
		app.args = cobra.NoArgs
	}
}

// WithWatchConfig enables watching and live-reloading the configuration file.
func WithWatchConfig() Option {
	return func(app *App) {
		app.watch = true
	}
}

// WithHealthCheckFunc sets a custom health check function that is started
// before the main run function.
func WithHealthCheckFunc(fn HealthCheckFunc) Option {
	return func(app *App) {
		app.healthCheckFunc = fn
	}
}

// WithOTel registers an OTel initializer that is called before the run function
// and shut down after it returns. Use otel.OTelOptions.Initializer() to create
// the init function. This keeps pkg/app free of OTel SDK dependencies.
func WithOTel(init OTelInitFunc) Option {
	return func(app *App) {
		app.otelInit = init
	}
}

// WithDefaultHealthCheckFunc enables the default health check server
// (with optional pprof) using HealthOptions defaults.
func WithDefaultHealthCheckFunc() Option {
	return WithHealthCheckFunc(func() error {
		go options.NewHealthOptions().Serve()
		return nil
	})
}

// NewApp creates a new application instance.
func NewApp(name string, shortDesc string, opts ...Option) *App {
	app := &App{
		name:      name,
		run:       func(context.Context) error { return nil },
		shortDesc: shortDesc,
	}

	for _, o := range opts {
		o(app)
	}

	app.buildCommand()
	return app
}

func (app *App) buildCommand() {
	cmd := &cobra.Command{
		Use:   formatBaseName(app.name),
		Short: app.shortDesc,
		Long:  app.description,
		RunE:  app.runCommand,
		Args:  app.args,
	}

	cmd.SilenceUsage = true
	cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		c.SilenceUsage = false
		return err
	})
	cmd.SilenceErrors = true

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.Flags().SortFlags = true

	var globalFS *pflag.FlagSet

	switch typed := app.options.(type) {
	case NamedFlagSetOptions:
		nfs := typed.Flags()
		globalFS = nfs.FlagSet("global")

		for _, name := range nfs.order {
			cmd.Flags().AddFlagSet(nfs.FlagSets[name])
		}

		SetUsageAndHelpFunc(cmd, nfs, 80)
	case FlagSetOptions:
		globalFS = cmd.PersistentFlags()
		typed.AddFlags(globalFS)
	default:
		globalFS = cmd.PersistentFlags()
	}

	version.AddFlags(globalFS)

	if !app.noConfig {
		AddConfigFlag(globalFS, app.name, app.watch)
	}

	app.cmd = cmd
}

// Run launches the application with OS signal handling for graceful shutdown.
// It installs SIGINT/SIGTERM handlers and injects a cancellable context into
// the cobra command, so the run function can observe ctx.Done() for shutdown.
// On error, it exits with status code 1.
func (app *App) Run() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app.cmd.SetContext(ctx)

	if err := app.cmd.ExecuteContext(ctx); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

func (app *App) runCommand(cmd *cobra.Command, args []string) error {
	version.PrintAndExitIfRequested()

	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	if app.options != nil {
		if err := viper.Unmarshal(app.options); err != nil {
			return err
		}

		if complete, ok := app.options.(interface{ Complete() error }); ok {
			if err := complete.Complete(); err != nil {
				return err
			}
		}

		if validate, ok := app.options.(interface{ Validate() error }); ok {
			if err := validate.Validate(); err != nil {
				return err
			}
		}
	}

	app.initializeLogger()

	if !app.silence {
		slog.Info("Starting application", "name", app.name, "version", version.Get().ToJSON())
		slog.Info("Golang settings", "GOGC", os.Getenv("GOGC"), "GOMAXPROCS", os.Getenv("GOMAXPROCS"), "GOTRACEBACK", os.Getenv("GOTRACEBACK"))
		if !app.noConfig {
			PrintConfig()
		} else if app.options != nil {
			PrintFlags(cmd.Flags(), func(format string, args ...any) {
				slog.Debug(format, args...)
			})
		}
	}

	if app.otelInit != nil {
		shutdown, err := app.otelInit(cmd.Context())
		if err != nil {
			return err
		}
		defer func() {
			if err := shutdown(context.Background()); err != nil {
				slog.Error("OTel shutdown failed", "error", err)
			}
		}()
	}

	if app.healthCheckFunc != nil {
		if err := app.healthCheckFunc(); err != nil {
			return err
		}
	}

	return app.run(cmd.Context())
}

// Command returns the underlying cobra.Command.
func (app *App) Command() *cobra.Command {
	return app.cmd
}

func formatBaseName(name string) string {
	if runtime.GOOS == "windows" {
		name = strings.ToLower(name)
		name = strings.TrimSuffix(name, ".exe")
	}
	return name
}

// initializeLogger sets up the logging system from viper configuration.
// Log options are read from the "log" key in the config (e.g., log.level, log.format).
func (app *App) initializeLogger() {
	logOptions := log.NewOptions()

	if err := viper.UnmarshalKey("log", logOptions); err != nil {
		slog.Warn("Failed to unmarshal log options from config, using defaults", "error", err)
	}

	log.Init(logOptions)
}
