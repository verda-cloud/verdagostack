// Package log provides a structured, leveled logging library for the
// verdagostack project, backed by go.uber.org/zap.
//
// It exposes both printf-style (*f) and structured key-value (*w) methods at
// every severity level (Debug, Info, Warn, Error, Panic, Fatal), a global
// logger with package-level convenience functions, context-aware field
// extraction via W(), and CLI flag integration through pflag.
//
// Quick start:
//
//	import "github.com/verda-cloud/verdagostack/pkg/log"
//
//	// Use package-level functions with the default global logger:
//	log.Infow("server starting", "port", 8080)
//	log.Infof("listening on :%d", 8080)
//
//	// Initialize with custom options (typically in main):
//	opts := log.NewOptions()
//	opts.Level = "debug"
//	log.Init(opts)
//	defer log.Sync()
//
//	// Pass the Logger interface for dependency injection:
//	svc := NewService(log.Default())
//
// Framework adapters for GORM and Kratos are available in sub-packages
// under pkg/log/adapter/. A no-op Logger for testing lives in pkg/log/empty/.
package log // import "github.com/verda-cloud/verdagostack/pkg/log"
